import { Routes, Route, Outlet } from "react-router-dom";
import Layout from "@/components/layouts/main-layout";
import Home from "@/pages/home";
import Settings from "@/pages/settings";
import Logs from "@/pages/logs";
import Login from "@/pages/login";
import { ProtectedRoute } from "@/components/protected-route";
import Magnet from "@/pages/magnet";

export default function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        element={
          <ProtectedRoute>
            <Layout>
              <Outlet />
            </Layout>
          </ProtectedRoute>
        }
      >
        <Route path="/" element={<Home />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/logs" element={<Logs />} />
        <Route path="/download" element={<Magnet />} />
        {/* 其他路由 */}
      </Route>
      {/* 404页面 */}
      <Route path="*" element={<div>404 Not Found</div>} />
    </Routes>
  );
}
