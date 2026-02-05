import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Profile = () => {
  const { user, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const [subscriptions, setSubscriptions] = useState([]);
  const [artists, setArtists] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    loadSubscriptions();
  }, [isAuthenticated, navigate]);

  const loadSubscriptions = async () => {
    try {
      setLoading(true);
      const subs = await api.getSubscriptions();
      const subscriptionsArray = Array.isArray(subs) ? subs : [];
      setSubscriptions(subscriptionsArray);
      
      // Load artist details for artist subscriptions
      const artistSubs = subscriptionsArray.filter(sub => sub && sub.type === 'artist');
      if (artistSubs.length > 0) {
        const artistPromises = artistSubs.map(sub => 
          api.getArtist(sub.artistId).catch(() => null)
        );
        const artistData = await Promise.all(artistPromises);
        setArtists(artistData.filter(a => a !== null));
      } else {
        setArtists([]);
      }
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju pretplata');
      setSubscriptions([]);
      setArtists([]);
    } finally {
      setLoading(false);
    }
  };

  const handleUnsubscribeArtist = async (artistId) => {
    if (!user) return;

    try {
      await api.unsubscribeFromArtist(artistId, user.id);
      setMessage('Uspešno ste se odjavili sa pretplate');
      setTimeout(() => setMessage(''), 3000);
      loadSubscriptions();
    } catch (err) {
      setError(err.message || 'Greška pri odjavi sa pretplate');
    }
  };

  const handleUnsubscribeGenre = async (genre) => {
    if (!user) return;

    try {
      await api.unsubscribeFromGenre(genre, user.id);
      setMessage('Uspešno ste se odjavili sa pretplate');
      setTimeout(() => setMessage(''), 3000);
      loadSubscriptions();
    } catch (err) {
      setError(err.message || 'Greška pri odjavi sa pretplate');
    }
  };

  const getArtistName = (artistId) => {
    const artist = artists.find(a => a.id === artistId);
    return artist ? artist.name : `Umetnik ${artistId}`;
  };

  if (!isAuthenticated) {
    return null;
  }

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  const artistSubscriptions = Array.isArray(subscriptions) 
    ? subscriptions.filter(sub => sub && sub.type === 'artist') 
    : [];
  const genreSubscriptions = Array.isArray(subscriptions) 
    ? subscriptions.filter(sub => sub && sub.type === 'genre') 
    : [];

  return (
    <div className="container">
      <div className="card">
        <h2>Moj Profil</h2>
        {user && (
          <div style={{ marginBottom: '20px' }}>
            <p><strong>Korisničko ime:</strong> {user.username}</p>
            <p><strong>Email:</strong> {user.email}</p>
            <p><strong>Uloga:</strong> {user.role === 'ADMIN' ? 'Administrator' : 'Korisnik'}</p>
          </div>
        )}
      </div>

      {message && (
        <div className="success" style={{ marginTop: '10px', marginBottom: '10px' }}>
          {message}
        </div>
      )}

      {error && (
        <div className="error" style={{ marginTop: '10px', marginBottom: '10px' }}>
          {error}
        </div>
      )}

      <div className="card">
        <h3>Pretplate na Umetnike ({artistSubscriptions.length})</h3>
        {artistSubscriptions.length === 0 ? (
          <p>Nemate pretplata na umetnike.</p>
        ) : (
          <div>
            {artistSubscriptions.map((sub) => (
              <div
                key={sub.id}
                className="list-item"
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '15px',
                  marginBottom: '10px',
                  border: '1px solid #ddd',
                  borderRadius: '5px'
                }}
              >
                <div>
                  <h4
                    style={{ margin: 0, cursor: 'pointer' }}
                    onClick={() => navigate(`/artists/${sub.artistId}`)}
                  >
                    {getArtistName(sub.artistId)}
                  </h4>
                  <p style={{ margin: '5px 0 0 0', fontSize: '0.9em', color: '#666' }}>
                    Pretplaćen od: {new Date(sub.createdAt).toLocaleDateString()}
                  </p>
                </div>
                <button
                  className="btn btn-secondary"
                  onClick={() => handleUnsubscribeArtist(sub.artistId)}
                  style={{ marginLeft: '10px' }}
                >
                  Otkaži pretplatu
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="card">
        <h3>Pretplate na Žanrove ({genreSubscriptions.length})</h3>
        {genreSubscriptions.length === 0 ? (
          <p>Nemate pretplata na žanrove.</p>
        ) : (
          <div>
            {genreSubscriptions.map((sub) => (
              <div
                key={sub.id}
                className="list-item"
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '15px',
                  marginBottom: '10px',
                  border: '1px solid #ddd',
                  borderRadius: '5px'
                }}
              >
                <div>
                  <h4 style={{ margin: 0 }}>
                    <span className="genre-tag" style={{ fontSize: '1.1em' }}>
                      {sub.genre}
                    </span>
                  </h4>
                  <p style={{ margin: '5px 0 0 0', fontSize: '0.9em', color: '#666' }}>
                    Pretplaćen od: {new Date(sub.createdAt).toLocaleDateString()}
                  </p>
                </div>
                <button
                  className="btn btn-secondary"
                  onClick={() => handleUnsubscribeGenre(sub.genre)}
                  style={{ marginLeft: '10px' }}
                >
                  Otkaži pretplatu
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Profile;
