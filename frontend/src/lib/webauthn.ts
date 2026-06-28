export function bufferToBase64url(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

export function base64urlToBuffer(str: string): ArrayBuffer {
  const base64 = str.replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), "=");
  const binary = atob(padded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes.buffer;
}

function decodePublicKeyCredentialCreationOptions(opts: any): CredentialCreationOptions {
  const raw = opts.publicKey || opts;
  const publicKey: any = { ...raw };
  publicKey.challenge = base64urlToBuffer(raw.challenge);
  publicKey.user.id = base64urlToBuffer(raw.user.id);
  if (raw.excludeCredentials) {
    publicKey.excludeCredentials = raw.excludeCredentials.map((cred: any) => ({
      ...cred,
      id: base64urlToBuffer(cred.id),
    }));
  }
  return { publicKey };
}

function decodeCredentialRequestOptions(opts: any): CredentialRequestOptions {
  const raw = opts.publicKey || opts;
  const publicKey: any = { ...raw };
  publicKey.challenge = base64urlToBuffer(raw.challenge);
  if (raw.allowCredentials) {
    publicKey.allowCredentials = raw.allowCredentials.map((cred: any) => ({
      ...cred,
      id: base64urlToBuffer(cred.id),
    }));
  }
  return { publicKey, mediation: opts.mediation };
}

function encodeCredential(cred: PublicKeyCredential): any {
  const response = cred.response as
    | AuthenticatorAttestationResponse
    | AuthenticatorAssertionResponse;
  const encoded: any = {
    id: cred.id,
    type: cred.type,
    rawId: bufferToBase64url(cred.rawId),
    response: {},
  };

  if ("attestationObject" in response) {
    const attResp = response as AuthenticatorAttestationResponse;
    encoded.response.attestationObject = bufferToBase64url(attResp.attestationObject);
    encoded.response.clientDataJSON = bufferToBase64url(attResp.clientDataJSON);
    if (attResp.getTransports) {
      encoded.response.transports = attResp.getTransports();
    }
  } else {
    const assertResp = response as AuthenticatorAssertionResponse;
    encoded.response.authenticatorData = bufferToBase64url(assertResp.authenticatorData);
    encoded.response.clientDataJSON = bufferToBase64url(assertResp.clientDataJSON);
    encoded.response.signature = bufferToBase64url(assertResp.signature);
    if (assertResp.userHandle) {
      encoded.response.userHandle = bufferToBase64url(assertResp.userHandle);
    }
  }

  return encoded;
}

export interface PasskeyRegisterBeginResponse {
  state_id: string;
  options: any;
}

export interface PasskeyLoginBeginResponse {
  state_id: string;
  options: any;
}

export async function startPasskeyRegistration(
  beginFn: () => Promise<PasskeyRegisterBeginResponse>,
  finishFn: (stateId: string, name: string, credential: any) => Promise<any>,
  name: string,
): Promise<any> {
  const { state_id, options } = await beginFn();
  const creationOptions = decodePublicKeyCredentialCreationOptions(options);
  const cred = await navigator.credentials.create(creationOptions);
  if (!cred) throw new Error("Passkey creation was cancelled");
  const encoded = encodeCredential(cred as PublicKeyCredential);
  return await finishFn(state_id, name, encoded);
}

export async function startPasskeyLogin(
  beginFn: () => Promise<PasskeyLoginBeginResponse>,
  finishFn: (stateId: string, credential: any) => Promise<any>,
): Promise<any> {
  const { state_id, options } = await beginFn();
  const requestOptions = decodeCredentialRequestOptions(options);
  const cred = await navigator.credentials.get(requestOptions);
  if (!cred) throw new Error("Passkey login was cancelled");
  const encoded = encodeCredential(cred as PublicKeyCredential);
  return await finishFn(state_id, encoded);
}
