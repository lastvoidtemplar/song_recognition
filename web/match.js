import { ApiHandler } from "./api_handler.js";

const micBtn = document.getElementById("mic");
const timer = document.getElementById("timer");

const playerDialog = document.getElementById("player-dialog");
const spinner = document.getElementById("spinner");
const songTitle = document.getElementById("song-title");
const player = document.getElementById("player");
const errorDialog = document.getElementById("error-dialog");

const apiUrl = window.location.origin;
const recordingTime = 20_000;
const timerStep = 40;

let interval;
let time = 0;

micBtn.onclick = async () => {
  micBtn.disabled = true;
  micBtn.classList.toggle("recording");

  time = 0;
  updateTimer();

  const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
  const mediaRecorder = new MediaRecorder(stream, { mimeType: "audio/webm" });
  setupMediaRecorder(mediaRecorder);
  mediaRecorder.start();

  interval = setInterval(() => {
    time += timerStep;
    updateTimer();

    if (time === recordingTime) {
      clearInterval(interval);
      micBtn.classList.toggle("recording");
      micBtn.disabled = false;

      clearDialog();
      playerDialog.showModal();
      // player.hidden = false;
      mediaRecorder.stop();
    }
  }, timerStep);
};

playerDialog.onclick = (e) => {
  if (e.target === playerDialog) {
    clearDialog();

    time = 0;
    updateTimer();

    playerDialog.close();
  }
};

function updateTimer() {
  const sec = Math.floor(time / 1000);
  const milliSec = Math.floor((time % 1000) / 10);
  timer.innerText = `${String(sec).padStart(2, "0")}.${String(
    milliSec
  ).padStart(2, "0")}`;
}

function setupMediaRecorder(mediaRecorder) {
  const audioChunks = [];

  mediaRecorder.ondataavailable = (e) => {
    if (e.data.size > 0) {
      audioChunks.push(e.data);
    }
  };

  mediaRecorder.onstop = async () => {
    console.log("recorder");

    const audioBlob = new Blob(audioChunks, { type: "audio/webm" });
    const formData = new FormData();
    formData.append("audio", audioBlob, `${crypto.randomUUID()}.webm`);

    const url = new URL("/match", apiUrl);
    const matchHandler = new ApiHandler(url, "post", formData);

    matchHandler.onLoading(() => {
      spinner.hidden = false;
    });

    matchHandler.onSuccess((data) => {
      spinner.hidden = true;

      songTitle.innerText = data.song_title;
      player.src = transformUrlToEmbedUrl(data.song_url);
      songTitle.hidden = false;
      player.hidden = false;
    });

    matchHandler.onError((_, err) => {
      spinner.hidden = true;
      errorDialog.innerText = err.error;
      errorDialog.hidden = false;
    });

    matchHandler.onFail((err) => {
      spinner.hidden = true;
      errorDialog.innerText = "Couldn`t connect to the server";
      errorDialog.hidden = false;
    });

    matchHandler.initiateFetch();
  };
}

function clearDialog() {
  errorDialog.innerText = "";
  errorDialog.hidden = true;
  songTitle.innerText = "";
  songTitle.hidden = true;
  player.hidden = true;
  player.src = ""
}

function transformUrlToEmbedUrl(songUrl) {
  const url = new URL(songUrl);
  return "https://www.youtube.com/embed/" + url.pathname.slice(1);
}
