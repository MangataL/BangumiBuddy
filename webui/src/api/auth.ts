import { http } from "@/api";

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
}

export interface AuthErrorResponse {
  message?: string;
  error_description?: string;
}

export const authApi = {
  login: async (username: string, password: string): Promise<AuthResponse> => {
    const params = new URLSearchParams();
    params.append("grant_type", "password");
    params.append("username", username);
    params.append("password", password);

    return http.post("/token", params, {
      skipAuth: true,
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
    });
  },

  refreshToken: async (refreshToken: string): Promise<AuthResponse> => {
    const params = new URLSearchParams();
    params.append("grant_type", "refresh_token");
    params.append("refresh_token", refreshToken);

    return http.post("/token", params, {
      skipAuth: true,
      headers: {
        "Content-Type": "application/x-www-form-urlencoded",
      },
    });
  },

  updateUser: async (username: string, password: string): Promise<void> => {
    return http.put("/user", { username, password });
  },
};

export default authApi;
