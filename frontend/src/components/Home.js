import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Home = () => {
  const { isAuthenticated, user } = useAuth();
  const [recommendations, setRecommendations] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (isAuthenticated && user?.id && user?.username !== 'admin') {
      fetchRecommendations();
    }
  }, [isAuthenticated, user]);

  const fetchRecommendations = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Koristi API Gateway endpoint umesto direktnog poziva na ratings-service
      const token = localStorage.getItem('token');
      const response = await fetch(`http://localhost:8081/api/ratings/recommendations?userId=${user.id}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });
      if (!response.ok) {
        throw new Error('Failed to fetch recommendations');
      }
      const data = await response.json();
      setRecommendations(data);
    } catch (err) {
      console.error('Error fetching recommendations:', err);
      setError('Unable to load recommendations');
    } finally {
      setLoading(false);
    }
  };

  const SongCard = ({ song, reason }) => (
    <div className="card" style={{ marginBottom: '15px', padding: '15px' }}>
      <h4>{song.name}</h4>
      <p><strong>Genre:</strong> {song.genre}</p>
      <p><strong>Duration:</strong> {Math.floor(song.duration / 60)}:{(song.duration % 60).toString().padStart(2, '0')}</p>
      <p><strong>Reason:</strong> {reason}</p>
      <Link to={`/songs/${song.songId}`} className="btn btn-primary btn-sm">
        View Song
      </Link>
    </div>
  );

  return (
    <div className="container">
      <div className="card">
        <h1>Dobrodošli u Music Streaming aplikaciju</h1>
        {isAuthenticated ? (
          <div>
            <p>Zdravo, {user.username}!</p>
            
            {/* Admin ne vidi preporuke */}
            {user?.username !== 'admin' && (
              <>
                {/* Recommendations Section */}
                {loading && <p>Loading recommendations...</p>}
                {error && <p style={{ color: 'red' }}>{error}</p>}
                
                {recommendations && (
                  <div style={{ marginTop: '30px' }}>
                    <h2>Personalized Recommendations</h2>
                    
                    {/* Songs from subscribed genres */}
                    {recommendations.subscribedGenreSongs && recommendations.subscribedGenreSongs.length > 0 && (
                      <div style={{ marginBottom: '30px' }}>
                        <h3>Based on your genre subscriptions</h3>
                        {recommendations.subscribedGenreSongs.map((song, index) => (
                          <SongCard key={index} song={song} reason={song.reason} />
                        ))}
                      </div>
                    )}
                    
                    {/* Top rated song from unsubscribed genre */}
                    {recommendations.topRatedSong && (
                      <div style={{ marginBottom: '30px' }}>
                        <h3>Discover something new</h3>
                        <SongCard song={recommendations.topRatedSong} reason={recommendations.topRatedSong.reason} />
                      </div>
                    )}
                    
                    {(!recommendations.subscribedGenreSongs || recommendations.subscribedGenreSongs.length === 0) && 
                     !recommendations.topRatedSong && (
                      <p>No recommendations available. Start subscribing to genres and rating songs to get personalized recommendations!</p>
                    )}
                  </div>
                )}
              </>
            )}
            
            {/* Admin vidi drugačiji sadržaj */}
            {user?.username === 'admin' && (
              <div style={{ marginTop: '30px' }}>
                <h2>Admin Panel</h2>
                <p>Welcome to admin dashboard. You have access to all administrative functions.</p>
              </div>
            )}
            
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
