new Vue({
  el: '#app',
  computed: {
    connected() {
      return window.Store.state.connected
    }
  }
})
