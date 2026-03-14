fn main() {
    println!("cargo:rerun-if-changed=../../go/grpc/proto/destination.proto");

    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .compile_protos(&["../../go/grpc/proto/destination.proto"], &["../../go/grpc/proto"])
        .expect("failed to compile destination.proto");
}
