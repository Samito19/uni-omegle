async function playLocalVideo() {
  try {
    const constraints = {
      'video': true,
      'audio': false
    }
    const videoStream = await navigator.mediaDevices.getUserMedia(constraints)
    const localVideoElement = document.querySelector("video#localVideo")
    localVideoElement.srcObject = videoStream;
  } catch (error) {
    console.error("Error getting video camera feed.")
    console.error(error)
  }
}

playLocalVideo()
