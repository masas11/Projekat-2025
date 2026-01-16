import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import api from '../services/api';

const ResetPassword = () => {
  const [searchParams] = useSearchParams();
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get('token');
    if (!token) {
      setError('Token za reset lozinke nije pronađen.');
    }
  }, [searchParams]);

  const validatePassword = (password) => {
    if (password.length < 8) {
      return 'Lozinka mora imati najmanje 8 karaktera';
    }
    if (!/[A-Z]/.test(password)) {
      return 'Lozinka mora sadržati najmanje jedno veliko slovo';
    }
    if (!/[0-9]/.test(password)) {
      return 'Lozinka mora sadržati najmanje jedan broj';
    }
    return null;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (password !== confirmPassword) {
      setError('Lozinke se ne poklapaju');
      return;
    }

    const passwordError = validatePassword(password);
    if (passwordError) {
      setError(passwordError);
      return;
    }

    const token = searchParams.get('token');
    if (!token) {
      setError('Token za reset lozinke nije pronađen.');
      return;
    }

    setLoading(true);

    try {
      await api.resetPassword(token, password);
      setSuccess('Lozinka je uspešno promenjena! Preusmeravanje na prijavu...');
      setTimeout(() => {
        navigate('/login');
      }, 2000);
    } catch (err) {
      setError(err.message || 'Greška pri resetovanju lozinke');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <div className="card">
        <h2>Reset lozinke</h2>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Nova lozinka:</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
            <small style={{ color: '#666', fontSize: '12px' }}>
              Lozinka mora imati najmanje 8 karaktera, jedno veliko slovo i jedan broj
            </small>
          </div>
          <div className="form-group">
            <label>Potvrdi novu lozinku:</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              required
            />
          </div>
          {error && <div className="error">{error}</div>}
          {success && <div className="success">{success}</div>}
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Resetovanje...' : 'Resetuj lozinku'}
          </button>
        </form>
      </div>
    </div>
  );
};

export default ResetPassword;
