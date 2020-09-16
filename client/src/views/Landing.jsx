import React, { useState } from 'react'
import { useEffect } from 'react'
import axios from 'axios'
import CopyToClipboard from 'react-copy-to-clipboard'
import { toast, ToastContainer } from 'react-toastify'
import Loader from 'react-loader-spinner'
import { ReactComponent as DownloadIcon } from '../assets/copy.svg'
import 'react-toastify/dist/ReactToastify.css'
import 'react-loader-spinner/dist/loader/css/react-spinner-loader.css'

const platforms_icons = {
  'deezer': require('../assets/logos/deezer.png'),
  'spotify': require('../assets/logos/spotify.png')
}

const BASE_URL = `https://zoove.herokuapp.com`

function pad(num, size) {
  let s = Math.floor(num) + '';
  while (s.length < size) {
    s = '0' + s;
  }
  return s;
}


function ConvertToMusicDuration(duration) {
  let hour = 0
  let minute = 0
  let seconds = 0

  let toSecs = duration / 1000
  minute = toSecs / 60
  seconds = toSecs % 60
  if (minute >= 60) {
    hour = minute / 60
    minute += minute / 60
  }

  return `${Math.floor(minute)}:${pad(seconds)}`
}

class Landing extends React.Component {

  state = {
    loading: false,
    tracks: [],
    srcTrack: '',
    clicked: false,
    isError: false,
    isDismissed: false,
    copied: false
  }

  async componentDidMount() {
  }

  SetSrcTrack(e) {
    this.setState({ srcTrack: e.target.value })
  }

  async FetchData() {
    try {
      this.setState({ loading: true })
      const { data: { data: reponse } } = await axios.get(`${BASE_URL}/api/v1/search?track=${this.state.srcTrack}`)
      const all = reponse.filter(x => x)
      this.setState({ loading: false, tracks: all })
    } catch (error) {
      console.log(`Error fetching data`)
    }
  }


  render() {
    return (
      <div className="flex bg-purple-1000 h-screen w-screen flex-col flex-1">
        <div>
          <header className="flex flex-row mx-2 xl:mx-112">
            <h1 className="text-6xl text-yellow-700 sm:text-3xl"> ZOOVE</h1>
          </header>
        </div>

        <div className="flex flex-col flex-1 xl:mx-80">
          <div className="flex flex-col">
            <h1 className="mx-2 w-9/12 font-bold text-4xl text-purple-1100 sm:text-2xl">Find the URL of a song on different platforms from one URL</h1>

            {this.state.isDismissed ? <></> : <div className="bg-teal-100 border-t-4 border-teal-500 rounded-b text-teal-900 px-4 py-3 shadow-md mr-20 ml-2" role="alert">
              <div className="flex">
                <div className="py-1"><svg className="fill-current h-6 w-6 text-red-500" role="button" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" onClick={() => this.setState({ isDismissed: !this.state.isDismissed })}>
                  <title>Close</title><path d="M14.348 14.849a1.2 1.2 0 0 1-1.697 0L10 11.819l-2.651 3.029a1.2 1.2 0 1 1-1.697-1.697l2.758-3.15-2.759-3.152a1.2 1.2 0 1 1 1.697-1.697L10 8.183l2.651-3.031a1.2 1.2 0 1 1 1.697 1.697l-2.758 3.152 2.758 3.15a1.2 1.2 0 0 1 0 1.698z" /></svg></div>
                <div>
                  <p className="font-bold">Things to note</p>
                  <p className="text-sm">Please note that only Spotify and Deezer are supported for now.
                  Apple Music support landing soon and Audiomack would be added
                    once API access request has been granted.</p>
                </div>
              </div>
            </div>}

            <div className="flex mt-5 ml-2">
              <input type="search" name="search" id="songsearch" placeholder="Paste song URL here."
                className="bg-purple-1000 focus:outline-none text-purple-1500 w-11/12 border-yellow-700 border border-l-0" onChange={e => this.SetSrcTrack(e)} />
              <button type="button" className="bg-yellow-700 text-white font-bold py-1 px-4 border border-yellow-700 rounded mb-2 mx-2 md:mr-32 mt-2" onClick={async e => {
                await this.FetchData()
              }}>Search</button>
            </div>


            {this.state.isError ? <span>An error occured and it's not you, it's Zoove. Please try again.</span> : ''}
            {this.state.tracks.length == 0 ? <span className="justify-center text-center text-purple-500 mt-10">An empty void..</span> : (this.state.loading ? <span className="text-center justify-center text-red-700">Loading</span> : this.state.tracks.map((x, y) => {
              return (
                <div className={`flex flex-col mx-10 mt-5 divide-y-4 divide-red-300 sm:mx-2 border-t-2 border-b-2 border-r-2 border-l-2 rounded-lg`} key={y} style={{ borderColor: '#a45ea8' }}>
                  <div className="flex flex-row justify-between my-2 ml-2">
                    <div className="flex flex-row overflow-hidden flex-1">
                      <img src={x?.cover}
                        alt="alty" srcSet="" style={{ width: '60px', height: '60px' }} className="rounded" />

                      <div className="flex flex-col">
                        <div className="flex flex-row ml-2">
                          <span className="text-purple-1700 mt-1 sm:w-32 md:w-9/12 xl:w-104 flex-no-wrap whitespace-no-wrap overflow-hidden " style={{ textOverflow: 'ellipsis' }}>{x?.title}</span>
                          {x?.explicit ? <abbr title="explicit" style={{
                            fontFamily: `Barlow,sans-sefif`, whiteSpace: 'pre',
                            display: 'block',
                            border: '2px solid red',
                            fontSize: '10px',
                            padding: '0 4px',
                            height: '16px',
                            fontWeight: "bold",
                            lineHeight: '12px',
                            textAlign: 'center',
                            marginLeft: '3px',
                            textDecoration: 'none',
                            opacity: 0.7,
                            textTransform: 'uppercase',
                            letterSpacing: '0.6px',
                            color: 'white',
                            borderColor: 'white'
                            ,
                            boxSizing: 'border-box'
                          }} className="mt-2">E</abbr> : ''}
                          <img src={platforms_icons[x?.platform]} alt="Image of platform it belongs to" style={{ height: '30px', width: '30px', marginLeft: '5px', borderRadius: '50%', }} />
                          <CopyToClipboard text={x?.url} onCopy={() => {
                            // this.setState({copied: true})
                            toast('Copied to clipboard')
                          }}>
                            {/* <span className="ml-2 text-red-400 w-full text-xs mt-2">Copy link</span> */}
                            {/* <DownloadIcon /> */}
                            <img src={require('../assets/copy.svg')} alt="Copy to Clipboard Image" style={{ height: '16px', width: '16px', marginTop: '8px', marginLeft: '2px' }} />
                          </CopyToClipboard>
                          {/* <img src={DownloadIcon} alt="" /> */}
                        </div>
                        <div className="flex flex-row">

                          <span className="ml-2 text-purple-200 sm:w-56 lg:w-11/12 md:w-56 whitespace-no-wrap block overflow-hidden" style={{ textOverflow: 'ellipsis' }}>{x?.artistes.join()}</span>
                          <ToastContainer />
                        </div>
                      </div>
                    </div>
                    <span className="mt-2 text-purple-1700 mr-2">{ConvertToMusicDuration(x?.duration)}</span>
                  </div>
                </div>
              )
            }))}
          </div>
        </div>
        <div>
          <footer>
            <div className="flex flex-col justify-center">
              <span className="italic mb-3 text-purple-1100 text-center mt-10">The media assets on this page are all copyright of their various owners.</span>
              <span className="text-center text-yellow-700">Zoove {new Date().getFullYear()}</span>
            </div>
          </footer>
        </div>
      </div>
    )
  }
}

// function Landing() {

//   useEffect(() => {
//     const fetchResult = async () => {
//       setLoading(true)
//       try {
//         const url = encodeURI(srcTrack)
//         const { data: { data: results } } = await axios.get(`${BASE_URL}/api/v1/search?track=${url}`)
//         const filt = results.filter(x => x)
//         setTracks(filt)
//       } catch (error) {
//         setIsError(true)
//       }
//       setLoading(false)
//     }
//     fetchResult()
//   }, [qrl])

//   // useEffect(() => {
//   //   setLoading(!loading)
//   // })
//   useEffect(() => {
//     document.title = "Zoove | One Link, Multiple Platforms"
//   })
//   const [isDismissed, setIsDismissed] = useState(false)

//   return (

//   )
// }

export default Landing

// return (
//   <div className="flex bg-purple-1000 flex-col h-screen w-screen flex-wrap">
// <header className="flex flex-row xl:mx-32 xl:justify-start md:justify-start mx-2 justify-center">
//   <div className="flex justify-center text-center">
//     <h1 className="text-5xl text-yellow-700 xl:mx-32">ZOOVE</h1>
//   </div>
// </header>
// <div className="flex flex-1 lg:mx-0 sm:mx-0 xl:mx-64 mt-10 flex-col flex-wrap">

//   <div className="flex w-4/6 mx-2 xl:mx-0">
//     <h1 className="text-purple-1100 font-bold sm: lg:mx-0 text-4xl"
//       style={{ fontFamily: `Barlow,sans-serif` }} >
//       Find music on multiple platforms the URL using the link from one</h1>
//   </div>
//   {/* <div>
//     <span className="flex-wrap w-4/6 flex mx-2 text-purple-1600 italic">Only Deezer and Spotify are supported now. Apple Music support coming soon.</span>
//   </div> */}
  // <div className="flex flex-row bg-purple-1000 border-b-4 justify-between lg:mx-2 mt-8 mx-2 focus:shadow-outline focus:bg-blue-300 border-yellow-700">
  //   <input type="search" name="search" id="songsearch" placeholder="Paste song URL here."
  //     style={{ fontFamily: `'Barlow'`, width: '90%' }}
  //     className="bg-purple-1000 focus:outline-none text-purple-1300" />
  //   <button className="bg-yellow-700 hover:bg-blue-700 text-white font-bold py-1 px-4 border border-yellow-700 rounded mb-2">Search</button>
  // </div>

  // {[1, 2, 3].map((x, y) => {
  //   return (
  //     <div className={`flex flex-col mx-10 mt-5 divide-y-4 divide-red-300 sm:mx-2 border-t-2 border-b-2 border-r-2 border-l-2 rounded-lg`} key={y} style={{ borderColor: '#a45ea8' }}>
  //       <div className="flex flex-row justify-between my-2 ml-2">
  //         <div className="flex flex-row overflow-hidden flex-1">
  //           <img src="https://static.highsnobiety.com/thumbor/RSyUUMRuA6AWUJLRJ3g2UoN_qlw=/fit-in/1000x600/smart/static.highsnobiety.com/wp-content/uploads/2017/04/15163510/adidas-yeezy-guide-wave-runner-main-2.jpg"
  //             alt="alty" srcSet="" style={{ width: '60px', height: '60px' }} className="rounded" />

  //           <div className="flex flex-col">
  //             <div className="flex flex-row ml-2">
  //               <span className="text-purple-1700 mt-1" style={{ textOverflow: 'ellipsis' }}>Burgundy</span>
  //               <abbr title="explicit" style={{
  //                 fontFamily: `Barlow,sans-sefif`, whiteSpace: 'pre',
  //                 display: 'block',
  //                 border: '2px solid red',
  //                 fontSize: '10px',
  //                 padding: '0 4px',
  //                 height: '16px',
  //                 fontWeight: "bold",
  //                 lineHeight: '12px',
  //                 textAlign: 'center',
  //                 marginLeft: '6px',
  //                 textDecoration: 'none',
  //                 opacity: 0.7,
  //                 textTransform: 'uppercase',
  //                 letterSpacing: '0.6px',
  //                 color: 'white',
  //                 borderColor: 'white'
  //                 ,
  //                 boxSizing: 'border-box'
  //               }} className="mt-2">E</abbr>
  //             </div>
  //             <span className="ml-2 text-purple-200 sm:w-56 lg:w-11/12 md:w-56 whitespace-no-wrap block overflow-hidden" style={{ textOverflow: 'ellipsis' }}>Earl Sweatshirt, Vince Staples, Domo Genesis, Syd</span>
  //           </div>
  //         </div>
  //         <span className="mt-2 text-purple-1700 mr-2">4:32</span>
  //       </div>
  //     </div>
  //   )
  // })}

//   <footer className="justify-center text-center mt-5 border-t-2 border-yellow-700">
//     <div className="flex flex-row justify-center">
//       <span className="mx-4 text-purple-1100">&#169;Zoove</span>
//     </div>
//   </footer>
// </div>
// </div >
// )