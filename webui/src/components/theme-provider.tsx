import * as React from "react";

type Theme = "light" | "dark" | "system";

interface ThemeProviderProps extends React.HTMLAttributes<HTMLElement> {
  children: React.ReactNode;
  attribute?: string;
  defaultTheme?: Theme;
  enableSystem?: boolean;
}

interface ThemeContextProps {
  theme?: Theme;
  setTheme: (theme: Theme) => void;
}

const ThemeContext = React.createContext<ThemeContextProps>({
  theme: undefined,
  setTheme: () => {},
});

function ThemeProvider({
  children,
  attribute = "class",
  defaultTheme = "system",
  enableSystem = true,
}: ThemeProviderProps) {
  const [theme, setThemeState] = React.useState<Theme>(() => {
    return ThemeStorage.getTheme(defaultTheme);
  });

  React.useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const root = window.document.documentElement;

    function updateTheme(theme: Theme) {
      if (!theme) return;

      const isSystem = theme === "system";
      const nextTheme = isSystem ? getSystemTheme() : theme;
      setThemeState(nextTheme);

      if (attribute === "class") {
        root.classList.remove("light", "dark");
        if (nextTheme !== "system") {
          root.classList.add(nextTheme);
        }
      } else {
        root.setAttribute("data-theme", nextTheme);
      }
    }

    updateTheme(theme);

    if (enableSystem) {
      const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
      const handleChange = () => {
        if (theme === "system") {
          updateTheme(theme);
        }
      };
      mediaQuery.addEventListener("change", handleChange);
      return () => mediaQuery.removeEventListener("change", handleChange);
    }
  }, [theme, attribute, enableSystem]);

  const setTheme = React.useCallback((theme: Theme) => {
    ThemeStorage.setTheme(theme);
    setThemeState(theme);
  }, []);

  return (
    <ThemeContext.Provider value={{ theme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

function useTheme() {
  return React.useContext(ThemeContext);
}

function getSystemTheme(): Theme {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export { ThemeProvider, useTheme };

class ThemeStorage {
  static setTheme(theme: Theme) {
    sessionStorage.setItem("theme", theme);
  }

  static getTheme(defaultTheme: Theme): Theme {
    return (sessionStorage.getItem("theme") as Theme) || defaultTheme;
  }
}
