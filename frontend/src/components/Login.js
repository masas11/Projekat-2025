import React, { useState, useEffect } from 'react';
import { useNavigate, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';

const Login = () => {
  const [step, setStep] = useState(1); // 1: credentials, 2: OTP
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [otp, setOtp] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    // Check if redirected from register with a message
    if (location.state?.message) {
      setSuccess(location.state.message);
      // Clear the state to prevent showing message on refresh
      window.history.replaceState({}, document.title);
    }
  }, [location]);

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
    <div style={{ 
      minHeight: 'calc(100vh - 80px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '40px 20px',
      background: 'linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%)'
    }}>
      <div className="card" style={{ 
        maxWidth: '480px', 
        width: '100%',
        background: 'linear-gradient(135deg, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.9) 100%)',
        backdropFilter: 'blur(10px)',
        boxShadow: '0 20px 60px rgba(0,0,0,0.15)',
        border: '1px solid rgba(255,255,255,0.3)'
      }}>
        <div style={{ textAlign: 'center', marginBottom: '40px' }}>
          <div style={{
            width: '80px',
            height: '80px',
            borderRadius: '20px',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '40px',
            margin: '0 auto 20px',
            boxShadow: '0 8px 20px rgba(102, 126, 234, 0.3)'
          }}>
            🎵
          </div>
          <h2 style={{ 
            fontSize: '32px', 
            fontWeight: '700',
            marginBottom: '8px',
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            backgroundClip: 'text'
          }}>
            Prijavi se
          </h2>
          <p style={{ color: '#666', fontSize: '14px' }}>Dobrodošli nazad!</p>
        </div>
        
        {success && (
          <div className="success" style={{ marginBottom: '20px' }}>
            <span>✅</span>
            <div>
              <p style={{ marginBottom: '8px' }}>{success}</p>
              <div style={{ marginTop: '10px', padding: '10px', backgroundColor: '#e3f2fd', borderRadius: '6px', border: '1px solid #90caf9' }}>
                <p style={{ fontSize: '12px', marginBottom: '5px', fontWeight: '500' }}>📧 MailHog Web UI:</p>
                <a 
                  href="http://localhost:8025" 
                  target="_blank" 
                  rel="noopener noreferrer"
                  style={{ 
                    color: '#1976d2', 
                    textDecoration: 'underline',
                    fontSize: '13px',
                    fontWeight: '500'
                  }}
                >
                  http://localhost:8025
                </a>
              </div>
            </div>
          </div>
        )}
        
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
            {error && (
              <div className="error">
                <span>⚠️</span>
                <span>{error}</span>
              </div>
            )}
            <button type="submit" className="btn btn-primary" disabled={loading} style={{ width: '100%', marginTop: '10px' }}>
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
              <small style={{ color: '#666', fontSize: '12px', display: 'block', marginTop: '5px' }}>
                Proverite MailHog (http://localhost:8025) gde vidite OTP kod
              </small>
            </div>
            {error && (
              <div className="error">
                <span>⚠️</span>
                <span>{error}</span>
              </div>
            )}
            <div style={{ display: 'flex', gap: '10px', marginTop: '10px' }}>
              <button type="submit" className="btn btn-primary" disabled={loading} style={{ flex: 1 }}>
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
        <div style={{ marginTop: '20px', textAlign: 'center', fontSize: '14px', color: '#666' }}>
          <div style={{ marginBottom: '10px' }}>
            <Link to="/forgot-password" style={{ color: '#667eea', textDecoration: 'none', fontWeight: '500' }}>
              Zaboravljena lozinka?
            </Link>
          </div>
          <div>
            <Link to="/recover-account" style={{ color: '#667eea', textDecoration: 'none', fontWeight: '500' }}>
              Povraćaj naloga (Magic Link)
            </Link>
          </div>
          <div style={{ marginTop: '15px', paddingTop: '15px', borderTop: '1px solid #e0e0e0' }}>
            Nemate nalog? <Link to="/register" style={{ color: '#667eea', textDecoration: 'none', fontWeight: '500' }}>Registrujte se</Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;
