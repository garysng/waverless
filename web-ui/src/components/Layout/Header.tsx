import { UserOutlined, LogoutOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { logout } from '@/utils/auth';
import './Header.css';

interface HeaderProps {
  title?: string;
}

const Header = ({ title = 'Dashboard' }: HeaderProps) => {
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header className="header">
      <h1 className="page-title">{title}</h1>
      <div className="user-info">
        <div className="user-avatar">
          <UserOutlined />
        </div>
        <span>Admin</span>
        <button className="btn-icon" onClick={handleLogout} title="Logout">
          <LogoutOutlined />
        </button>
      </div>
    </header>
  );
};

export default Header;
