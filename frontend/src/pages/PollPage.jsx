import { useState, useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'
import { getPoll, votePoll, getWsUrl } from '../api.js'

function PollPage() {
  const { id } = useParams()
  const [poll, setPoll] = useState(null)
  const [counts, setCounts] = useState({})
  const [voted, setVoted] = useState(false)
  const [error, setError] = useState('')
  const socketRef = useRef(null)

  useEffect(() => {
    async function loadPoll() {
      const data = await getPoll(id)
      if (data.error) {
        setError(data.error)
        return
      }
      setPoll(data)
      setCounts(data.counts || {})
    }

    loadPoll()

    if (localStorage.getItem('voted_' + id)) {
      setVoted(true)
    }

    const socket = new WebSocket(getWsUrl(id))
    socketRef.current = socket

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.counts) {
        setCounts(data.counts)
      }
    }

    return () => {
      socket.close()
    }
  }, [id])

  async function handleVote(optionId) {
    const data = await votePoll(id, optionId)
    if (data.error) {
      setError(data.error)
      return
    }
    localStorage.setItem('voted_' + id, 'true')
    setVoted(true)
  }

  if (error) {
    return <p className="error">{error}</p>
  }

  if (!poll) {
    return <p>Loading...</p>
  }

  const totalVotes = Object.values(counts).reduce((sum, c) => sum + Number(c), 0)

  return (
    <div className="form-box">
      <h2>{poll.question}</h2>

      <div className="share-link">
        Share this link
        <input type="text" readOnly value={window.location.href} />
      </div>

      {poll.options.map((opt) => {
        const count = Number(counts[opt.id] || 0)
        const percent = totalVotes > 0 ? Math.round((count / totalVotes) * 100) : 0

        return (
          <div key={opt.id} className="option-row">
            <div className="option-top">
              {!voted ? (
                <button onClick={() => handleVote(opt.id)}>{opt.text}</button>
              ) : (
                <span>{opt.text}</span>
              )}
              <span>{count} votes ({percent}%)</span>
            </div>
            <div className="bar-bg">
              <div className="bar-fill" style={{ width: percent + '%' }}></div>
            </div>
          </div>
        )
      })}
    </div>
  )
}

export default PollPage
