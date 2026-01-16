import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Login = () => {
  const [step, setStep] = useState(1); // 1: credentials, 2: OTP
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [otp, setOtp] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleRequestOTP = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await api.requestOTP({ username, password });
      setStep(2);
    } catch (err) {
      setError(err.message || 'Greška pri prijavljivanju');
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOTP = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const response = await api.verifyOTP(username, otp);
      login(response, response.token);
      navigate('/');
    } catch (err) {
      setError(err.message || 'Nevažeći OTP kod');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <div className="card">
        <h2>Prijavi se</h2>
        
        {step === 1 ? (
          <form onSubmit={handleRequestOTP}>
            <div className="form-group">
              <label>Korisničko ime:</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label>Lozinka:</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            {error && <div className="error">{error}</div>}
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? 'Slanje...' : 'Zatraži OTP'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleVerifyOTP}>
            <div className="form-group">
              <label>OTP kod (proverite email):</label>
              <input
                type="text"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                placeholder="123456"
                required
              />
              <small style={{ color: '#666', fontSize: '12px' }}>
                Proverite konzolu servera gde vidite OTP kod
              </small>
            </div>
            {error && <div className="error">{error}</div>}
            <div style={{ display: 'flex', gap: '10px' }}>
              <button type="submit" className="btn btn-primary" disabled={loading}>
                {loading ? 'Verifikacija...' : 'Verifikuj OTP'}
              </button>
              <button
                type="button"
                className="btn btn-secondary"
                onClick={() => {
                  setStep(1);
                  setOtp('');
                  setError('');
                }}
              >
                Nazad
              </button>
            </div>
          </form>
        )}
        <div style={{ marginTop: '15px', textAlign: 'center', fontSize: '14px' }}>
          <Link to="/forgot-password" style={{ marginRight: '15px' }}>
            Zaboravljena lozinka?
          </Link>
          <Link to="/recover-account">Povraćaj naloga (Magic Link)</Link>
        </div>
      </div>
    </div>
  );
};

export default Login;
