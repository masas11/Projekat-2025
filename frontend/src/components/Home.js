import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Home = () => {
  const { isAuthenticated, user } = useAuth();

  return (
    <div className="container">
      <div className="card">
        <h1>Dobrodošli u Music Streaming aplikaciju</h1>
        {isAuthenticated ? (
          <div>
            <p>Zdravo, {user.username}!</p>
            <p>Istražite našu kolekciju muzike:</p>
            <div style={{ marginTop: '20px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
              <Link to="/artists" className="btn btn-primary">Izvođači</Link>
              <Link to="/albums" className="btn btn-primary">Albumi</Link>
              <Link to="/songs" className="btn btn-primary">Pesme</Link>
              <Link to={`/notifications?userId=${user.id}`} className="btn btn-primary">Notifikacije</Link>
            </div>
          </div>
        ) : (
          <div>
            <p>Molimo vas da se prijavite ili registrujete da biste pristupili aplikaciji.</p>
            <div style={{ marginTop: '20px', display: 'flex', gap: '10px' }}>
              <Link to="/login" className="btn btn-primary">Prijavi se</Link>
              <Link to="/register" className="btn btn-secondary">Registruj se</Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Home;
