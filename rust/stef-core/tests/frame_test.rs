use stef_core::{Compression, FrameDecoder, FrameEncoder, FrameFlags, MemReaderWriter};

fn test_last_frame_and_continue(compression: Compression) {
    let mut encoder = FrameEncoder::default();
    let mut buf = MemReaderWriter::default();
    encoder.init(&mut buf, compression).unwrap();

    let write = "hello".repeat(10).into_bytes();
    encoder.write(&write).unwrap();
    encoder.close_frame().unwrap();

    // reset read pointer
    buf.set_position(0);

    let mut decoder = FrameDecoder::default();
    decoder.init(buf.clone(), compression).unwrap();
    decoder.next().unwrap();

    let mut read = vec![0u8; write.len()];
    let n = decoder.read(&mut read).unwrap();
    assert_eq!(n, write.len());
    assert_eq!(read, write);

    let err = decoder.read(&mut read).unwrap_err();
    assert!(err.to_string().contains("end of frame"));
}

#[test]
fn test_last_frame_and_continue_none() {
    test_last_frame_and_continue(Compression::None);
}

#[test]
fn test_last_frame_and_continue_zstd() {
    test_last_frame_and_continue(Compression::Zstd);
}
