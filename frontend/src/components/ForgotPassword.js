import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

const ForgotPassword = () => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
      await api.requestPasswordReset(email);
      setSuccess('Ako email postoji, link za reset lozinke je poslat na vašu email adresu.');
    } catch (err) {
      setError(err.message || 'Greška pri slanju email-a');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <div className="card">
        <h2>Zaboravljena lozinka</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Email adresa:</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="unesite@email.com"
            />
            <small style={{ color: '#666', fontSize: '12px' }}>
              Unesite email adresu povezanu sa vašim nalogom
            </small>
          </div>
          {error && <div className="error">{error}</div>}
          {success && <div className="success">{success}</div>}
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Slanje...' : 'Pošalji link za reset'}
          </button>
          <div style={{ marginTop: '15px', textAlign: 'center' }}>
            <Link to="/login">Nazad na prijavu</Link>
          </div>
        </form>
      </div>
    </div>
  );
};

export default ForgotPassword;
