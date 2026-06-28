import { describe, it, expect } from "vitest";
import { bufferToBase64url, base64urlToBuffer } from "./webauthn";

describe("base64url", () => {
  it("encodes and decodes an empty buffer", () => {
    const buf = new ArrayBuffer(0);
    const encoded = bufferToBase64url(buf);
    expect(encoded).toBe("");
    const decoded = base64urlToBuffer(encoded);
    expect(decoded.byteLength).toBe(0);
  });

  it("encodes and decodes a simple byte sequence", () => {
    const buf = new Uint8Array([0x00, 0x01, 0x02, 0xfe, 0xff]).buffer;
    const encoded = bufferToBase64url(buf);
    expect(encoded).toBe("AAEC_v8");
    const decoded = base64urlToBuffer(encoded);
    expect(new Uint8Array(decoded)).toEqual(new Uint8Array([0x00, 0x01, 0x02, 0xfe, 0xff]));
  });

  it("encodes and decodes bytes that need padding", () => {
    const buf = new Uint8Array([0x61, 0x62, 0x63]).buffer; // "abc"
    const encoded = bufferToBase64url(buf);
    expect(encoded).toBe("YWJj"); // no padding in base64url
    const decoded = base64urlToBuffer(encoded);
    expect(new Uint8Array(decoded)).toEqual(new Uint8Array([0x61, 0x62, 0x63]));
  });

  it("handles binary data with high bytes", () => {
    const buf = new Uint8Array(Array.from({ length: 256 }, (_, i) => i)).buffer;
    const encoded = bufferToBase64url(buf);
    const decoded = base64urlToBuffer(encoded);
    expect(new Uint8Array(decoded)).toEqual(
      new Uint8Array(Array.from({ length: 256 }, (_, i) => i)),
    );
  });

  it("roundtrips large random data", () => {
    const data = new Uint8Array(1024);
    crypto.getRandomValues(data);
    const encoded = bufferToBase64url(data.buffer);
    const decoded = base64urlToBuffer(encoded);
    expect(new Uint8Array(decoded)).toEqual(data);
  });
});
