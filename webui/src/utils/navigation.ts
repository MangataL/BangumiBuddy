import { TokenService } from "./token";

// 创建一个导航服务，用于在非React上下文中进行导航
class NavigationService {
  private static navigate: (to: string) => void;

  static init(navigateFunction: (to: string) => void) {
    this.navigate = navigateFunction;
  }

  static navigateToLogin(redirectPath?: string) {
    TokenService.clearTokens()
    const path = `/login${
      redirectPath ? `?redirect=${encodeURIComponent(redirectPath)}` : ""
    }`;
    if (this.navigate) {
      this.navigate(path);
    } else {
      // 降级方案：如果navigate未初始化，使用window.location
      window.location.href = path;
    }
  }
}

export default NavigationService;
