import { NavLink, useLocation } from 'react-router-dom';
import {
  DashboardOutlined,
  RocketOutlined,
  CloudServerOutlined,
  UnorderedListOutlined,
  DatabaseOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';
import './Sidebar.css';

interface SidebarProps {
  collapsed: boolean;
  onCollapse: (collapsed: boolean) => void;
}

const menuItems = [
  { path: '/dashboard', icon: <DashboardOutlined />, label: 'Dashboard' },
  { path: '/serverless', icon: <RocketOutlined />, label: 'Serverless' },
  { path: '/endpoints', icon: <CloudServerOutlined />, label: 'Endpoints' },
  { path: '/tasks', icon: <UnorderedListOutlined />, label: 'Tasks' },
  { path: '/specs', icon: <DatabaseOutlined />, label: 'Specs' },
];

const Sidebar = ({ collapsed, onCollapse }: SidebarProps) => {
  const location = useLocation();

  return (
    <div className={`sidebar ${collapsed ? 'collapsed' : ''}`}>
      <div className="logo">
        <RocketOutlined style={{ fontSize: 20, color: '#1da1f2' }} />
        {!collapsed && <span>Waverless</span>}
      </div>
      <nav className="nav-menu">
        {menuItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              `nav-item ${isActive || location.pathname.startsWith(item.path) ? 'active' : ''}`
            }
          >
            {item.icon}
            {!collapsed && <span>{item.label}</span>}
          </NavLink>
        ))}
      </nav>
      <div
        className="nav-item collapse-btn"
        onClick={() => onCollapse(!collapsed)}
        style={{ position: 'absolute', bottom: 12, width: '100%' }}
      >
        {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
        {!collapsed && <span>Collapse</span>}
      </div>
    </div>
  );
};

export default Sidebar;
