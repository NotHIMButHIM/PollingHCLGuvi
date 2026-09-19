import { Routes, Route, Link, useNavigate } from 'react-router-dom'
import { useState, useEffect } from 'react'
import Login from './pages/Login.jsx'
import Signup from './pages/Signup.jsx'
import CreatePoll from './pages/CreatePoll.jsx'
import PollPage from './pages/PollPage.jsx'

function App() {
  const [token, setToken] = useState(localStorage.getItem('token') || '')
  const navigate = useNavigate()

  useEffect(() => {
    if (token) {
      localStorage.setItem('token', token)
    }
  }, [token])

  function handleLogout() {
    localStorage.removeItem('token')
    setToken('')
    navigate('/')
  }

  return (
    <div className="container">
      <nav className="navbar">
        <Link to="/" className="brand">Live Polling</Link>
        <div className="nav-links">
          {token ? (
            <button onClick={handleLogout}>Logout</button>
          ) : (
            <>
              <Link to="/login">Login</Link>
              <Link to="/signup">Signup</Link>
            </>
          )}
        </div>
      </nav>

      <Routes>
        <Route path="/" element={<CreatePoll token={token} />} />
        <Route path="/login" element={<Login setToken={setToken} />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/poll/:id" element={<PollPage />} />
      </Routes>
    </div>
  )
}

export default App
