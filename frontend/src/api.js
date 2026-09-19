const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export async function signup(name, email, password) {
  const res = await fetch(BASE_URL + '/api/auth/signup', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, email, password })
  })
  return res.json()
}

export async function login(email, password) {
  const res = await fetch(BASE_URL + '/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  })
  return res.json()
}

export async function createPoll(question, options, token) {
  const res = await fetch(BASE_URL + '/api/polls', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: 'Bearer ' + token
    },
    body: JSON.stringify({ question, options })
  })
  return res.json()
}

export async function getPoll(id) {
  const res = await fetch(BASE_URL + '/api/polls/' + id)
  return res.json()
}

export async function votePoll(id, optionId) {
  const res = await fetch(BASE_URL + '/api/polls/' + id + '/vote', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ optionId })
  })
  return res.json()
}

export function getWsUrl(id) {
  const wsBase = BASE_URL.replace('http', 'ws')
  return wsBase + '/ws/polls/' + id
}
