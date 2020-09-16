import React from 'react'
import qs from 'query-string'
import axios from 'axios'


class DeezerAuth extends React.Component {
  state = { code: '' }

  async componentDidMount() {
    try {
      const APP_ID = '422202'
      const SECRET_KEY = 'e39a5133a1ab818926e848e5695e644c'
      const DEEZER_BASE_URL = `https://connect.deezer.com/oauth/access_token.php`

      const qz = qs.parse(window.location.search)
      const code = qz?.code
      const URL = `${DEEZER_BASE_URL}?app_id=${APP_ID}&secret=${SECRET_KEY}&code=${code}`

      const { data } = await axios.get(URL, { headers: { 'Content-Type': 'text/plain' } })
      console.log(data)
      this.setState({ code })
    } catch (error) {
      console.log(`Error in componentDidMount`)
      console.log(error)
      // probably set some error state here
    }
  }
  render() {
    const { code } = this.state
    return (
      <div>
        <div>
          <span>I see you...{code}</span>
        </div>
      </div>
    )
  }
}

export default DeezerAuth