var conn = new WebSocket("ws://localhost:8080/ws");
const configuration = { iceServers: [{ urls: "stun:stun.l.google.com:19302" }] }
const peerConnection = new RTCPeerConnection(configuration)
const remoteVideoElement = document.querySelector("video#remoteVideo")
const localVideoElement = document.querySelector("video#localVideo")
var peerRole
var makingOffer = false;
var ignoreOffer = false;


async function playLocalVideo() {
  try {
    const constraints = {
      'video': true,
      'audio': true
    }

    const mediaStream = await navigator.mediaDevices.getUserMedia(constraints)

    for (const track of mediaStream.getTracks()) {
      console.log("added track")
      peerConnection.addTrack(track, mediaStream)
    }

    localVideoElement.srcObject = mediaStream;
  } catch (error) {
    console.error("Error with media stream feed.")
    console.error(error)
  }
}
playLocalVideo()
peerConnection.addEventListener('track', async (event) => {
  const [remoteStream] = event.streams;
  remoteVideoElement.srcObject = remoteStream;
});

peerConnection.onnegotiationneeded = async () => {
  try {
    makingOffer = true;
    await peerConnection.setLocalDescription();
    conn.send(`{"description": ${JSON.stringify(peerConnection.localDescription)}}`)
  } catch (err) {
    console.error(err);
  } finally {
    makingOffer = false;
  }
};

peerConnection.onicecandidate = ({ candidate }) => {
  if (!candidate)
    return
  conn.send(JSON.stringify(candidate));
}

// async function makeCall() {
//   console.log("Make Call Button Clicked")
//   const offer = await peerConnection.createOffer()
//   await peerConnection.setLocalDescription(offer)
//   conn.send(JSON.stringify(offer))
// }

// peerConnection.ontrack = ({ track, streams }) => {
//   console.log("recevied trackQ")
//   track.onunmute = () => {
//     console.log("recevied trackQ")
//     if (remoteVideoElement.srcObject) {
//       return;
//     }
//     remoteVideoElement.srcObject = streams[0];
//   };
// };

conn.onmessage = async function(evt) {
  var messages = evt.data.split('\n');
  for (var i = 0; i < messages.length; i++) {
    var message = JSON.parse(messages[i])
    if (!message) {
      continue
    }
    try {
      if (message["peerType"]) {
        peerRole = message["peerType"]
        console.log("peerRole: " + peerRole)
      } else if (message["description"]) {
        console.log("Received description!")
        const description = message["description"]
        console.log(description["type"] === "offer", peerConnection.signalingState, makingOffer, peerRole === "impolite")
        const offerCollision =
          description["type"] === "offer" &&
          (makingOffer || peerConnection.signalingState !== "stable");

        ignoreOffer = peerRole === "impolite" && offerCollision;

        console.log(ignoreOffer, peerRole, offerCollision)
        if (ignoreOffer) {
          return;
        }
        await peerConnection.setRemoteDescription(description);
        if (description.type === "offer") {
          await peerConnection.setLocalDescription();
          conn.send(`{"description": ${JSON.stringify(peerConnection.localDescription)}}`)
        }
      } else if (message["candidate"]) {
        const candidate = message
        try {
          console.log("Added candidate!")
          await peerConnection.addIceCandidate(candidate);
        } catch (err) {
          if (!ignoreOffer) {
            throw err;
          }
        }
      }
    } catch (err) {
      console.error(err)
    }
  };

  conn.onclose = function(evt) {
    console.log("Conn closed")
  };

}
