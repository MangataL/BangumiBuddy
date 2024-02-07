// 提供全局的认证状态管理
class AuthService {
  private static logoutCallback: (() => void) | null = null;

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
}

export default AuthService;
