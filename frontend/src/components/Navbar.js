import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Navbar = () => {
  const { user, logout, isAdmin } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await api.logout();
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      logout();
      navigate('/login');
    }
  };

  return (
    <nav className="navbar">
      <div className="navbar-content">
        <div>
          <Link to="/">Music Streaming</Link>
          <Link to="/artists">Izvođači</Link>
          <Link to="/albums">Albumi</Link>
          <Link to="/songs">Pesme</Link>
          {user && <Link to={`/notifications?userId=${user.id}`}>Notifikacije</Link>}
          {user && <Link to="/change-password">Promena lozinke</Link>}
        </div>
        <div className="navbar-user">
          {user ? (
            <>
              <span>
                {user.username}
                {isAdmin() && <span className="badge badge-admin">Admin</span>}
                {!isAdmin() && <span className="badge badge-user">Korisnik</span>}
              </span>
              <button className="btn btn-secondary" onClick={handleLogout}>
                Odjavi se
              </button>
            </>
          ) : (
            <>
              <Link to="/login">Prijavi se</Link>
              <Link to="/register">Registruj se</Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
