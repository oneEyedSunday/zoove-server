import React from 'react'
import axios from 'axios'

class Signup extends React.Component {
  state = {}

  async Deezer() {
    try {
      const DEEZER_BASE_URL = `https://connect.deezer.com/oauth/auth.php`
      const APP_ID = '422202'
      const SECRET_KEY = 'e39a5133a1ab818926e848e5695e644c'
      const REDIRECT_URI = encodeURIComponent(`https://a69e0edb0774.ngrok.io/deezer/oauth`)
      const PERMISSIONS = 'basic_access,email,offline_access'
      const URL = `${DEEZER_BASE_URL}?app_id=${APP_ID}&redirect_uri=${REDIRECT_URI}&perms=${PERMISSIONS}`
      // await axios.get(URL)
      window.location = URL
      return
    } catch (error) {
      console.log(`Error signinng up with Deezer`)
      console.log(error)
      return
    }
  }
  render() {
    return (
      <div className="container mx-auto">
        <div>
          <div className="bg-blue-700 mt-12 flex flex-row">
            <div className="flex flex-col flex-1">
              <span className="md:ml-40 md:my-10">Parscorum</span>
              <span className="ml-12">Sign up and share your music record!</span>
            </div>

            <div className="bg-red-400 flex-1 ">
              <span className="flex justify-center">Hiiii</span>
              <button className="bg-purple-600 text-white py-2 rounded mx-10" onClick={async e => await this.Deezer()}>Signup with Deezer</button>
              <button className="bg-pink-500 text-white py-2 rounded mx-10">Signup with Spotify</button>
              <button className="bg-purple-600 text-white py-2 rounded mx-10 mt-2">Signup with YTMusic</button>
            </div>
          </div>
        </div>
      </div>
    )
  }
}

export default Signup