import { jwtDecode } from "jwt-decode";

interface TokenPayload {
  exp: number;
  iat: number;
  sub: string;
}

interface Tokens {
  accessToken: string;
  refreshToken: string;
}

const TOKEN_KEY = "auth_tokens";
const SESSION_TOKEN_KEY = "session_auth_tokens";

export class TokenService {
  static storeTokens(tokens: Tokens, remember: boolean) {
    if (remember) {
      localStorage.setItem(TOKEN_KEY, JSON.stringify(tokens));
    } else {
      sessionStorage.setItem(SESSION_TOKEN_KEY, JSON.stringify(tokens));
    }
  }

  static getTokens(): Tokens | null {
    const localTokens = localStorage.getItem(TOKEN_KEY);
    const sessionTokens = sessionStorage.getItem(SESSION_TOKEN_KEY);

    return localTokens
      ? JSON.parse(localTokens)
      : sessionTokens
      ? JSON.parse(sessionTokens)
      : null;
  }

  static clearTokens() {
    localStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(SESSION_TOKEN_KEY);
  }

  static isTokenExpired(token: string): boolean {
    try {
      const decoded = jwtDecode<TokenPayload>(token);
      return decoded.exp * 1000 < Date.now();
    } catch {
      return true;
    }
  }

  static getAccessToken(): string | null {
    const tokens = this.getTokens();
    return tokens?.accessToken || null;
  }

  static getRefreshToken(): string | null {
    const tokens = this.getTokens();
    return tokens?.refreshToken || null;
  }
}
