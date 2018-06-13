window.Store = new Vuex.Store({
  state: {
    connected: false,
    state: {}
  },
  mutations: {
    setConnected: (state, value) => {
      state.connected = value
    },
    setState: (state, value) => {
      state.state = value
    }
  }
})

window.Store.setConnecting = true

const wsURL = (window.location.protocol === 'https' ? 'wss://' : 'ws://') +
  window.location.host + '/ws'
const socket = new ReconnectingWebSocket(wsURL, null, {
  debug: true,
  reconnectInterval: 3000
})

socket.addEventListener('open', () => {
  window.Store.commit('setConnected', true)
  console.log('WS connection opened')
})
socket.addEventListener('message', (event) => {
  window.Store.commit('setState', JSON.parse(event.data))
})
socket.addEventListener('close', () => {
  window.Store.commit('setConnected', false)
  console.log('WS connection closed')
})
