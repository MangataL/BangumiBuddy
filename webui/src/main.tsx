import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { BrowserRouter } from "react-router-dom";
import AppRoutes from "./router";
import "./globals.css";

const root = document.getElementById("root");
if (!root) throw new Error("Root element not found");

createRoot(root).render(
  <React.StrictMode>
    <BrowserRouter>
      <App>
        <AppRoutes />
      </App>
    </BrowserRouter>
  </React.StrictMode>
);
