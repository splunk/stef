use std::{
    io::{BufRead, BufReader},
    net::SocketAddr,
    path::PathBuf,
    process::{Command, Output, Stdio},
    sync::{Arc, mpsc},
    thread,
    time::{Duration, Instant},
};

use stef_core::{ChunkWriter, schema::{Schema, Struct, WireSchema}};
use stef_grpc::{
    Callbacks, Client, ClientCallbacks, ClientSchema, ClientSettings, ServerSettings, StreamServer, proto, types::Logger,
};
use tokio::net::TcpListener as TokioTcpListener;
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
    let listener = TokioTcpListener::bind("127.0.0.1:0").await.unwrap();
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

fn repo_root() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR"))
        .parent()
        .and_then(|p| p.parent())
        .expect("workspace layout changed")
        .to_path_buf()
}

fn output_with_timeout(mut child: std::process::Child, timeout: Duration) -> Output {
    let deadline = Instant::now() + timeout;
    loop {
        match child.try_wait().expect("failed waiting for child process") {
            Some(_) => return child.wait_with_output().expect("failed reading child output"),
            None if Instant::now() < deadline => thread::sleep(Duration::from_millis(25)),
            None => {
                let _ = child.kill();
                return child.wait_with_output().expect("failed reading timed out child output");
            }
        }
    }
}

fn wait_for_port(addr: SocketAddr, timeout: Duration) {
    let deadline = Instant::now() + timeout;
    while Instant::now() < deadline {
        if std::net::TcpStream::connect(addr).is_ok() {
            return;
        }
        thread::sleep(Duration::from_millis(20));
    }
    panic!("timed out waiting for server on {addr}");
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
async fn test_interop_go_client_to_rust_server() {
    // This test verifies cross-language streaming from the Go client into the Rust server.
    // The Rust server checks that the exact chunk bytes arrive and sends an ack back to Go.
    let expected_chunk: Vec<u8> = vec![0x01, 0x02, 0x03, 0xaa, 0xbb, 0xcc];
    let (chunk_tx, chunk_rx) = mpsc::channel::<Vec<u8>>();

    let listener = TokioTcpListener::bind("127.0.0.1:0").await.expect("failed to bind rust server");
    let addr = listener.local_addr().expect("failed to get local addr");

    let stream_server = StreamServer::new(ServerSettings {
        logger: Some(Arc::new(TestLogger)),
        server_schema: make_wire_schema("TestStruct", 2),
        max_dict_bytes: 1024,
        callbacks: Callbacks {
            on_stream: Some(Arc::new(move |mut reader, stream| {
                let mut got = vec![0u8; expected_chunk.len()];
                let mut filled = 0usize;
                while filled < got.len() {
                    let n = reader.read(&mut got[filled..])?;
                    if n == 0 {
                        return Err(tonic::Status::internal("unexpected EOF while reading chunk"));
                    }
                    filled += n;
                }
                if got != expected_chunk {
                    return Err(tonic::Status::invalid_argument(format!(
                        "unexpected bytes: got={got:02x?} want={expected_chunk:02x?}"
                    )));
                }
                tokio::runtime::Handle::current().block_on(async {
                    stream
                        .send_data_response(proto::StefDataResponse {
                            ack_record_id: 1,
                            ..Default::default()
                        })
                        .await
                })?;
                let _ = chunk_tx.send(got);
                Ok(())
            })),
        },
    });

    let (shutdown_tx, shutdown_rx) = tokio::sync::oneshot::channel::<()>();
    tokio::spawn(async move {
        Server::builder()
            .add_service(stream_server.service())
            .serve_with_incoming_shutdown(
                tokio_stream::wrappers::TcpListenerStream::new(listener),
                async move {
                    let _ = shutdown_rx.await;
                },
            )
            .await
            .expect("rust server failed");
    });

    tokio::task::spawn_blocking(move || wait_for_port(addr, Duration::from_secs(3)))
        .await
        .expect("wait_for_port task panicked");

    let go_dir = repo_root().join("go/grpc");
    let go_target = format!("localhost:{}", addr.port());
    let output = tokio::task::spawn_blocking(move || {
        Command::new("go")
            .current_dir(go_dir)
            .env("CGO_ENABLED", "0")
            .args([
                "run",
                "./cmd/interop",
                "client",
                "--target",
                &go_target,
                "--root-struct",
                "TestStruct",
                "--field-count",
                "2",
                "--header-hex",
                "010203",
                "--content-hex",
                "aabbcc",
                "--expect-ack",
                "1",
                "--timeout",
                "5s",
            ])
            .output()
            .expect("failed to run go interop client")
    })
    .await
    .expect("go client task panicked");

    assert!(
        output.status.success(),
        "go client failed\nstdout:\n{}\nstderr:\n{}",
        String::from_utf8_lossy(&output.stdout),
        String::from_utf8_lossy(&output.stderr)
    );

    let received = chunk_rx
        .recv_timeout(Duration::from_secs(5))
        .expect("rust server did not receive expected chunk from go client");
    assert_eq!(received, vec![0x01, 0x02, 0x03, 0xaa, 0xbb, 0xcc]);
    let _ = shutdown_tx.send(());
}

#[tokio::test]
async fn test_interop_rust_client_to_go_server() {
    // This test verifies the opposite direction: Rust client sends a chunk to the Go server.
    // The Go server validates bytes and returns ack 7, and the Rust client callback confirms that ack.
    let go_dir = repo_root().join("go/grpc");

    let mut child = Command::new("go")
        .current_dir(go_dir)
        .env("CGO_ENABLED", "0")
        .args([
            "run",
            "./cmd/interop",
            "server",
            "--listen",
            "127.0.0.1:0",
            "--root-struct",
            "TestStruct",
            "--field-count",
            "2",
            "--expected-chunk-hex",
            "102030aabbccdd",
            "--ack-id",
            "7",
            "--timeout",
            "8s",
        ])
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .expect("failed to start go interop server");

    let stdout = child.stdout.take().expect("missing go server stdout");
    let ready_line = tokio::task::spawn_blocking(move || {
        let mut reader = BufReader::new(stdout);
        let mut line = String::new();
        reader.read_line(&mut line).expect("failed to read go server ready line");
        line
    })
    .await
    .expect("ready-line task panicked");

    let listen_addr = ready_line
        .strip_prefix("READY ")
        .map(str::trim)
        .expect("go server did not print ready line")
        .to_string();

    let wait_addr = listen_addr.parse::<SocketAddr>().expect("invalid go server socket addr");
    tokio::task::spawn_blocking(move || wait_for_port(wait_addr, Duration::from_secs(3)))
        .await
        .expect("wait_for_port task panicked");

    let (ack_tx, ack_rx) = tokio::sync::oneshot::channel::<u64>();
    let ack_tx = Arc::new(std::sync::Mutex::new(Some(ack_tx)));
    let mut client = Client::new(ClientSettings {
        logger: Some(Arc::new(TestLogger)),
        grpc_endpoint: format!("http://{listen_addr}"),
        client_schema: ClientSchema {
            root_struct_name: "TestStruct".into(),
            wire_schema: make_wire_schema("TestStruct", 2),
        },
        callbacks: ClientCallbacks {
            on_disconnect: Arc::new(|_| {}),
            on_ack: Arc::new(move |ack_id| {
                if let Some(tx) = ack_tx.lock().expect("poisoned ack lock").take() {
                    let _ = tx.send(ack_id);
                }
                Ok(())
            }),
        },
    })
    .await
    .expect("failed to create rust client");

    let (mut writer, _opts) = tokio::time::timeout(Duration::from_secs(3), client.connect())
        .await
        .expect("rust client connect timed out")
        .expect("rust client connect failed");

    writer
        .write_chunk(&[0x10, 0x20, 0x30], &[0xaa, 0xbb, 0xcc, 0xdd])
        .expect("rust client failed to write chunk");

    let ack_id = tokio::time::timeout(Duration::from_secs(5), ack_rx)
        .await
        .expect("timed out waiting for go server ack")
        .expect("ack channel dropped unexpectedly");
    assert_eq!(ack_id, 7, "rust client received unexpected ack id");

    tokio::time::timeout(Duration::from_secs(2), client.disconnect())
        .await
        .expect("rust client disconnect timed out")
        .expect("rust client disconnect failed");

    let output = tokio::task::spawn_blocking(move || output_with_timeout(child, Duration::from_secs(5)))
        .await
        .expect("go server wait task panicked");
    assert!(
        output.status.success(),
        "go server failed\nstdout:\n{}\nstderr:\n{}",
        String::from_utf8_lossy(&output.stdout),
        String::from_utf8_lossy(&output.stderr)
    );
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
