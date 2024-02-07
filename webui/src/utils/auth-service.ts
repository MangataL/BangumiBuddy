// 提供全局的认证状态管理
class AuthService {
  private static logoutCallback: (() => void) | null = null;
  private static getAccessTokenFn: (() => Promise<string | null>) | null = null;
  // 注册登出回调函数
  static registerLogoutCallback(callback: () => void) {
    this.logoutCallback = callback;
  }

  // 取消注册登出回调函数
  static unregisterLogoutCallback() {
    this.logoutCallback = null;
  }

  // 触发未授权状态（401响应）
  static handleUnauthorized() {
    if (this.logoutCallback) {
      this.logoutCallback();
    }
  }

  static setGetAccessToken(getAccessToken: () => Promise<string | null>) {
    this.getAccessTokenFn = getAccessToken;
  }

  // 获取访问令牌
  static async getAccessToken(): Promise<string | null> {
    if (this.getAccessTokenFn) {
      return this.getAccessTokenFn();
    }
    return null;
  }
}

export default AuthService;
