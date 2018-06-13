const st = window.Store.state
Vue.component('dashboard', {
  template: `<div class="container">
    <h1>RFID bridge</h1>
    <div class="row">
      <div class="col-lg-6">
        <h2>Device connectivity</h2>
        <table class="table">
          <tr>
            <td style="width: 50%">Modem address</td>
            <td style="width: 50%"><code>{{ modemAddress }}</code></td>
          </tr>
          <tr>
            <td>Modem available</td>
            <td>
              <span v-if="modemPing >= 0" class="badge badge-success">{{ modemPing }}ms</span>
              <span v-else class="badge badge-danger">Down</span>
            </td>
          </tr>
          <tr>
            <td>RFID3 client running</td>
            <td>
              <span v-if="modemRunning" class="badge badge-success">Yes</span>
              <span v-else class="badge badge-danger">No</span>
            </td>
          </tr>
          <tr>
            <td>Database engine</td>
            <td>
              <span v-if="databaseRunning" class="badge badge-success">Yes</span>
              <span v-else class="badge badge-danger">No</span>
            </td>
          </tr>
        </table>
      </div>
      <div class="col-lg-6">
        <h2>Synchronization</h2>
        <table class="table">
          <tr>
            <td style="width: 50%">Upstream address</td>
            <td style="width: 50%"><code>{{ upstreamAddress }}</code></td>
          </tr>
          <tr>
            <td>Engine running</td>
            <td>
              <span v-if="upstreamRunning" class="badge badge-success">Yes</span>
              <span v-else class="badge badge-danger">No</span>
            </td>
          </tr>
          <tr>
            <td>Local record count</td>
            <td>
              <code>{{ localCount }}</code>
            </td>
          </tr>
          <tr>
            <td>Remote record count</td>
            <td>
              <code>{{ upstreamCount }}</code>
            </td>
          </tr>
        </table>
      </div>
    </div>
    <div class="row">
      <div class="col-lg-12">
        <h2>Most recent readings</h2>
        <table class="table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Tag ID</th>
              <th>Antenna</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in readouts">
              <td>{{ item.timestamp | format }}</td>
              <td>{{ item.tag_id }}</td>
              <td>{{ item.antenna }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>`,
  computed: {
    modemAddress () {
      return st.state && st.state.modem && st.state.modem.address
    },
    modemPing () {
      const startValue = st.state && st.state.modem && st.state.modem.ping
      if (!startValue || startValue === -1) {
        return startValue
      }

      return Number.parseFloat(startValue / 1000000).toFixed(2)
    },
    modemRunning () {
      return st.state && st.state.modem && st.state.modem.running
    },
    databaseRunning () {
      return st.state && st.state.database && st.state.database.running
    },
    localCount () {
      return st.state && st.state.database && st.state.database.count
    },
    upstreamRunning () {
      return st.state && st.state.upstream && st.state.upstream.connected
    },
    upstreamAddress () {
      return st.state && st.state.upstream && st.state.upstream.address
    },
    upstreamCount () {
      return st.state && st.state.upstream && st.state.upstream.count
    },
    readouts () {
      return st.state && st.state.readouts && st.state.readouts.length > 0 && st.state.readouts.reverse() || []
    }
  },
  filters: {
    format(value) {
      return moment(value).format("DD/MM/YY HH:mm:ss")
    }
  }
})