import { useState } from 'react';
import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import Sidebar from './components/Layout/Sidebar';
import Header from './components/Layout/Header';
import DashboardPage from './pages/Dashboard';
import ServerlessPage from './pages/Serverless';
import EndpointsPage from './pages/Endpoints';
import EndpointDetailPage from './pages/EndpointDetail';
import TasksPage from './pages/Tasks';
import SpecsPage from './pages/Specs';
import LoginPage from './pages/Login';
import { isAuthenticated } from './utils/auth';
import './styles/portal.css';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  if (!isAuthenticated()) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
}

// Page title mapping
const pageTitles: Record<string, string> = {
  '/dashboard': 'Dashboard',
  '/serverless': 'Serverless',
  '/endpoints': 'Endpoints',
  '/tasks': 'Tasks',
  '/specs': 'Hardware Specs',
};

function App() {
  const location = useLocation();
  const isLoginPage = location.pathname === '/login';
  const [collapsed, setCollapsed] = useState(false);

  if (isLoginPage) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
      </Routes>
    );
  }

  // Get page title
  const getPageTitle = () => {
    for (const [path, title] of Object.entries(pageTitles)) {
      if (location.pathname.startsWith(path)) return title;
    }
    return 'Waverless';
  };

  return (
    <div className="app">
      <Sidebar collapsed={collapsed} onCollapse={setCollapsed} />
      <div className="main-content" style={{ marginLeft: collapsed ? 60 : 200 }}>
        <Header title={getPageTitle()} />
        <div className="content">
          <Routes>
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
            <Route path="/serverless" element={<ProtectedRoute><ServerlessPage /></ProtectedRoute>} />
            <Route path="/endpoints" element={<ProtectedRoute><EndpointsPage /></ProtectedRoute>} />
            <Route path="/endpoints/:name" element={<ProtectedRoute><EndpointDetailPage /></ProtectedRoute>} />
            <Route path="/tasks" element={<ProtectedRoute><TasksPage /></ProtectedRoute>} />
            <Route path="/specs" element={<ProtectedRoute><SpecsPage /></ProtectedRoute>} />
            {/* Legacy routes redirect */}
            <Route path="/apps" element={<Navigate to="/endpoints" replace />} />
            <Route path="/apps/:name" element={<Navigate to="/endpoints/:name" replace />} />
            <Route path="/deploy" element={<Navigate to="/serverless" replace />} />
          </Routes>
        </div>
      </div>
    </div>
  );
}

export default App;
