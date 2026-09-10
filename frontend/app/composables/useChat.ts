export interface ChatMessage {
  id: string
  channel_id: string
  organization_id: string
  project_id: string
  user_id: string
  message: string
  message_type: string
  created_at: string
  user?: { id: string; email: string; full_name: string }
}

export function useChat(projectId: Ref<string | null>) {
  const { get, post } = useApi()
  const config = useRuntimeConfig()
  const token = useState<string | null>('auth:token', () => null)
  const orgId = useState<string | null>('org:org_id', () => null)

  const messages = ref<ChatMessage[]>([])
  const loading = ref(false)
  const sending = ref(false)
  const channelName = ref('')

  let ws: WebSocket | null = null
  let pollTimer: any = null
  let reconnectTimer: any = null

  function getEffectiveIds() {
    const pid = projectId.value
    const oid = orgId.value || (import.meta.client ? localStorage.getItem('auth:org_id') : null)
    const tok = token.value || (import.meta.client ? localStorage.getItem('auth:token') : null)
    return { pid, oid, tok }
  }

  async function fetchMessages() {
    const { pid } = getEffectiveIds()
    if (!pid) return
    try {
      const data = await get<{ channel: any; messages: ChatMessage[] }>(`/api/v1/projects/${pid}/chat`)
      messages.value = data.messages || []
      channelName.value = data.channel?.name || 'Chat'
    } catch (e) {
      // silent
    }
  }

  async function sendMessage(text: string) {
    const { pid } = getEffectiveIds()
    if (!pid || !text.trim()) return
    sending.value = true
    try {
      const msg = await post<ChatMessage>(`/api/v1/projects/${pid}/chat`, { message: text.trim(), message_type: 'text' })
      messages.value.push(msg)
      return msg
    } finally {
      sending.value = false
    }
  }

  function connectWS() {
    if (!import.meta.client) return
    const { pid, tok } = getEffectiveIds()
    if (!pid || !tok) return
    // close existing
    if (ws) { try { ws.close() } catch {} ; ws = null }
    const base = (config.public.apiBaseUrl as string).replace(/^http/, 'ws')
    const url = `${base}/api/v1/projects/${pid}/chat/ws?token=${encodeURIComponent(tok)}`
    try {
      ws = new WebSocket(url)
      ws.onmessage = (ev) => {
        try {
          // bisa multi-json separated by \n
          const parts = ev.data.split('\n').filter(Boolean)
          for (const p of parts) {
            const envelope = JSON.parse(p)
            // envelope.data is the message, or envelope itself
            const msg: ChatMessage = envelope.data?.id ? envelope.data : envelope
            if (msg?.id && !messages.value.find(m => m.id === msg.id)) {
              messages.value.push(msg)
            }
          }
        } catch {}
      }
      ws.onclose = () => {
        // fallback to polling reconnect
        reconnectTimer = setTimeout(() => connectWS(), 3000)
      }
      ws.onerror = () => {
        try { ws?.close() } catch {}
      }
    } catch {
      // fallback polling
    }
  }

  function startPolling() {
    if (pollTimer) clearInterval(pollTimer)
    pollTimer = setInterval(fetchMessages, 3000)
  }

  function start() {
    fetchMessages()
    connectWS()
    startPolling()
  }

  function stop() {
    if (ws) { try { ws.close() } catch {} ; ws = null }
    if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  }

  watch(projectId, () => {
    messages.value = []
    stop()
    if (projectId.value) start()
  })

  onMounted(() => {
    if (projectId.value) start()
  })
  onUnmounted(stop)

  return { messages, loading, sending, channelName, fetchMessages, sendMessage, connectWS }
}
