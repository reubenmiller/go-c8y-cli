import React from "react"
// reference: https://www.gatsbyjs.com/docs/how-to/images-and-media/working-with-video/#embedding-hosted-videos-in-markdown
// https://asciinema.org/a/326455.js

const videoStyle = {
    display: 'flex',
    float: 'none',
    overflow: 'hidden',
    padding: 0,
    margin: '20px 0',
    flex: 1,
    justifyContent: 'space-around',
};

const Video = ({ videoSrcURL, videoTitle, width, height, ...props }) => (
  <div className="video" style={videoStyle}>
    <iframe
      src={videoSrcURL}
      title={videoTitle}
      allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
      frameBorder="0"
      scrolling="yes"
      webkitallowfullscreen="true"
      mozallowfullscreen="true"
      allowFullScreen
      style={{ width, height, overflow: 'hidden', margin: 0, border: 0, display: 'inline-block', float: 'none' }}
      {...props}
    />
  </div>
)
export default Video