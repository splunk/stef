use std::sync::Arc;

use stef_core::schema::{Schema, Struct, WireSchema};
use stef_grpc::{Callbacks, Client, ClientCallbacks, ClientSchema, ClientSettings, ServerSettings, StreamServer, types::Logger};
use tokio::net::TcpListener;
use tonic::transport::Server;

#[derive(Default)]
struct TestLogger;
impl Logger for TestLogger {
    fn debugf(&self, _msg: &str) {}
    fn errorf(&self, _msg: &str) {}
}

fn make_wire_schema(name: &str, field_count: usize) -> WireSchema {
    let mut s = Schema::default();
    let mut st = Struct { name: name.to_string(), is_root: true, ..Default::default() };
    for i in 0..field_count {
        st.fields.push(stef_core::schema::StructField { name: format!("f{i}"), ..Default::default() });
    }
    s.structs.insert(name.to_string(), st);
    WireSchema::new(&s, name)
}

async fn start_server(schema: WireSchema) -> String {
    let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();

    let stream_server = StreamServer::new(ServerSettings {
        logger: Some(Arc::new(TestLogger)),
        server_schema: schema,
        max_dict_bytes: 0,
        callbacks: Callbacks::default(),
    });

    tokio::spawn(async move {
        Server::builder()
            .add_service(stream_server.service())
            .serve_with_incoming(tokio_stream::wrappers::TcpListenerStream::new(listener))
            .await
            .unwrap();
    });

    format!("http://{}", addr)
}

#[tokio::test]
async fn test_schema_compatibility_exact() {
    let endpoint = start_server(make_wire_schema("TestStruct", 2)).await;

    let mut client = Client::new(ClientSettings {
        logger: Some(Arc::new(TestLogger)),
        grpc_endpoint: endpoint,
        client_schema: ClientSchema { root_struct_name: "TestStruct".into(), wire_schema: make_wire_schema("TestStruct", 2) },
        callbacks: ClientCallbacks::default(),
    })
    .await
    .unwrap();

    let (_writer, opts) = tokio::time::timeout(std::time::Duration::from_secs(3), client.connect())
        .await
        .expect("connect timed out")
        .unwrap();
    assert!(!opts.include_descriptor);

    tokio::time::timeout(std::time::Duration::from_secs(3), client.disconnect())
        .await
        .expect("disconnect timed out")
        .unwrap();
}

#[tokio::test]
async fn test_schema_compatibility_superset_requires_descriptor() {
    let endpoint = start_server(make_wire_schema("TestStruct", 3)).await;

    let mut client = Client::new(ClientSettings {
        logger: Some(Arc::new(TestLogger)),
        grpc_endpoint: endpoint,
        client_schema: ClientSchema { root_struct_name: "TestStruct".into(), wire_schema: make_wire_schema("TestStruct", 2) },
        callbacks: ClientCallbacks::default(),
    })
    .await
    .unwrap();

    let (_writer, opts) = tokio::time::timeout(std::time::Duration::from_secs(3), client.connect())
        .await
        .expect("connect timed out")
        .unwrap();
    assert!(opts.include_descriptor);
    assert!(opts.schema.is_some());

    tokio::time::timeout(std::time::Duration::from_secs(3), client.disconnect())
        .await
        .expect("disconnect timed out")
        .unwrap();
}
