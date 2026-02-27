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
        <div style={{ display: 'flex', alignItems: 'center', gap: '20px', flexWrap: 'wrap' }}>
          <Link to="/" style={{ fontSize: '20px', fontWeight: '600', textDecoration: 'none' }}>
            🎵 Music Streaming
          </Link>
          <Link to="/artists" style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
            Izvođači
          </Link>
          <Link to="/albums" style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
            Albumi
          </Link>
          <Link to="/songs" style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
            Pesme
          </Link>
          {user && (
            <>
              <Link to={`/notifications?userId=${user.id}`} style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
                Notifikacije
              </Link>
              <Link to="/activity-history" style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
                Istorija Aktivnosti
              </Link>
              {!isAdmin() && (
                <Link to="/analytics" style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
                  Analitike
                </Link>
              )}
              <Link to="/profile" style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
                Moj Profil
              </Link>
              <Link to="/change-password" style={{ textDecoration: 'none', transition: 'opacity 0.2s' }} onMouseEnter={(e) => e.target.style.opacity = '0.8'} onMouseLeave={(e) => e.target.style.opacity = '1'}>
                Promena lozinke
              </Link>
            </>
          )}
        </div>
        <div className="navbar-user">
          {user ? (
            <>
              <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <span style={{ fontWeight: '500' }}>{user.username}</span>
                {isAdmin() && <span className="badge badge-admin">Admin</span>}
                {!isAdmin() && <span className="badge badge-user">Korisnik</span>}
              </span>
              <button className="btn btn-secondary" onClick={handleLogout} style={{ padding: '8px 16px', fontSize: '14px' }}>
                Odjavi se
              </button>
            </>
          ) : (
            <>
              <Link to="/login" style={{ textDecoration: 'none', padding: '8px 16px', borderRadius: '6px', transition: 'background-color 0.2s' }} onMouseEnter={(e) => e.target.style.backgroundColor = 'rgba(255,255,255,0.1)'} onMouseLeave={(e) => e.target.style.backgroundColor = 'transparent'}>
                Prijavi se
              </Link>
              <Link to="/register" className="btn btn-primary" style={{ textDecoration: 'none', padding: '8px 16px', fontSize: '14px' }}>
                Registruj se
              </Link>
            </>
          )}
        </div>
      </div>
    </nav>
  );
};

export default Navbar;
