const inputAddSong = document.getElementById("input-add-song");
const btnAddSong = document.getElementById("btn-add-song");
const errorAddSong = document.getElementById("error-add-song");
const iframeSong = document.getElementById("iframe-song");

btnAddSong.addEventListener("click", async () => {
  const url = inputAddSong.value;

  const res = await fetch("/song", {
    method: "POST",
    body: JSON.stringify({
      song_url: url,
    }),
  });

  if (res.status !== 200) {
    const body = await res.text();
    errorAddSong.innerText = body;
  }
});

const btnRecorder = document.getElementById("btn-recorder");

btnRecorder.addEventListener("click", async () => {
  iframeSong.hidden = true
  const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
  const mediaRecorder = new MediaRecorder(stream, {mimeType: "audio/webm"});
  const audioChunks = [];

  mediaRecorder.ondataavailable = (e) => {
    if (e.data.size > 0) {
      audioChunks.push(e.data);
    }
  };

  mediaRecorder.onstop = async () => {
    const audioBlob = new Blob(audioChunks, {type: "audio/webm"});
    const formData = new FormData();
    formData.append("audio", audioBlob, `${crypto.randomUUID()}.webm`);

    const response = await fetch("/match", {
      method: "POST",
      body: formData,
    });

    btnRecorder.innerText = "Record"

    if (response.ok){
      showSong(await response.text())
    }
  };

  mediaRecorder.start();
  btnRecorder.innerText = "Recording...";

  setTimeout(() => {
    mediaRecorder.stop();
    btnRecorder.innerText = "Processing...";
  }, 20000);
});


function showSong(song) {
  const url = new URL(song)
  iframeSong.src = "https://www.youtube.com/embed/"+url.pathname.slice(1)
  iframeSong.hidden = false
}