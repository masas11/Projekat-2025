import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const ArtistDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuth();
  const [artist, setArtist] = useState(null);
  const [albums, setAlbums] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [subscriptionMessage, setSubscriptionMessage] = useState('');
  const [isSubscribing, setIsSubscribing] = useState(false);
  const [isSubscribed, setIsSubscribed] = useState(false);

  useEffect(() => {
    loadArtist();
    loadAlbums();
    if (isAuthenticated && user) {
      checkSubscription();
    }
  }, [id, isAuthenticated, user]);

  const loadArtist = async () => {
    try {
      const data = await api.getArtist(id);
      setArtist(data);
    } catch (err) {
      setError(err.message || 'Greška pri učitavanju izvođača');
    } finally {
      setLoading(false);
    }
  };

  const loadAlbums = async () => {
    try {
      const data = await api.getAlbumsByArtist(id);
      setAlbums(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Error loading albums:', err);
    }
  };

  const checkSubscription = async () => {
    if (!isAuthenticated || !user) return;

    try {
      const subscriptions = await api.getSubscriptions();
      const subscribed = subscriptions.some(
        sub => sub.type === 'artist' && sub.artistId === id
      );
      setIsSubscribed(subscribed);
    } catch (err) {
      console.error('Error checking subscription:', err);
    }
  };

  const handleSubscribe = async () => {
    if (!isAuthenticated || !user) {
      setError('Morate biti prijavljeni da biste se pretplatili na umetnika');
      return;
    }

    setIsSubscribing(true);
    setSubscriptionMessage('');
    setError('');

    try {
      if (isSubscribed) {
        await api.unsubscribeFromArtist(id, user.id);
        setIsSubscribed(false);
        setSubscriptionMessage('Uspešno ste se odjavili sa pretplate na ovog umetnika!');
      } else {
        await api.subscribeToArtist(id, user.id);
        setIsSubscribed(true);
        setSubscriptionMessage('Uspešno ste se pretplatili na ovog umetnika!');
      }
      setTimeout(() => setSubscriptionMessage(''), 3000);
    } catch (err) {
      if (err.message && err.message.includes('Already subscribed')) {
        setIsSubscribed(true);
        setError('Već ste pretplaćeni na ovog umetnika');
      } else {
        setError(err.message || 'Greška pri pretplati na umetnika');
      }
    } finally {
      setIsSubscribing(false);
    }
  };

  if (loading) {
    return <div className="container">Učitavanje...</div>;
  }

  if (error || !artist) {
    return (
      <div className="container">
        <div className="card">
          <div className="error">{error || 'Izvođač nije pronađen'}</div>
          <button className="btn btn-secondary" onClick={() => navigate('/artists')}>
            Nazad
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="card">
        <button className="btn btn-secondary" onClick={() => navigate('/artists')} style={{ marginBottom: '20px' }}>
          ← Nazad na izvođače
        </button>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
          <h2 style={{ margin: 0 }}>{artist.name}</h2>
          {isAuthenticated && (
            <button
              className={isSubscribed ? "btn btn-secondary" : "btn btn-primary"}
              onClick={handleSubscribe}
              disabled={isSubscribing}
              style={{ marginLeft: '20px' }}
            >
              {isSubscribing 
                ? (isSubscribed ? 'Odjavljivanje...' : 'Pretplaćivanje...') 
                : (isSubscribed ? '✓ Pretplaćen' : '🔔 Pretplati se')}
            </button>
          )}
        </div>
        {subscriptionMessage && (
          <div className="success" style={{ marginTop: '10px', marginBottom: '10px' }}>
            {subscriptionMessage}
          </div>
        )}
        {artist.biography && <p style={{ marginTop: '10px', marginBottom: '10px' }}>{artist.biography}</p>}
        {artist.genres && artist.genres.length > 0 && (
          <div style={{ marginTop: '10px', marginBottom: '20px' }}>
            <strong>Žanrovi: </strong>
            {artist.genres.map((genre, idx) => (
              <span key={idx} className="genre-tag">{genre}</span>
            ))}
          </div>
        )}
      </div>

      {albums.length > 0 && (
        <div className="card">
          <h3>Albumi</h3>
          {albums.map((album) => (
            <div
              key={album.id}
              className="list-item"
              onClick={() => navigate(`/albums/${album.id}`)}
            >
              <h4>{album.name}</h4>
              {album.genre && <span className="genre-tag">{album.genre}</span>}
              {album.releaseDate && <p style={{ marginTop: '5px' }}>Datum izdavanja: {new Date(album.releaseDate).toLocaleDateString()}</p>}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ArtistDetail;
