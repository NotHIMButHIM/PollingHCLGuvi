import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { createPoll } from '../api.js'

function CreatePoll({ token }) {
  const [question, setQuestion] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [error, setError] = useState('')
  const navigate = useNavigate()

  function updateOption(index, value) {
    const updated = [...options]
    updated[index] = value
    setOptions(updated)
  }

  function addOption() {
    setOptions([...options, ''])
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')

    const cleanedOptions = options.filter((o) => o.trim() !== '')
    if (cleanedOptions.length < 2) {
      setError('at least 2 options required')
      return
    }

    const data = await createPoll(question, cleanedOptions, token)
    if (data.error) {
      setError(data.error)
      return
    }

    navigate('/poll/' + data.id)
  }

  if (!token) {
    return (
      <div className="form-box">
        <h2>Create a Poll</h2>
        <p>You need to login first to create a poll.</p>
        <Link to="/login">Go to Login</Link>
      </div>
    )
  }

  return (
    <div className="form-box">
      <h2>Create a Poll</h2>
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Poll question"
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          required
        />
        {options.map((opt, index) => (
          <input
            key={index}
            type="text"
            placeholder={'Option ' + (index + 1)}
            value={opt}
            onChange={(e) => updateOption(index, e.target.value)}
          />
        ))}
        <button type="button" onClick={addOption}>
          Add Option
        </button>
        <button type="submit">Create Poll</button>
      </form>
      {error && <p className="error">{error}</p>}
    </div>
  )
}

export default CreatePoll
