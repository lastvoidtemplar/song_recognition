const micBtn = document.getElementById("mic");
const timer = document.getElementById("timer");
const playerDialog = document.getElementById("player-dialog")
const player = document.getElementById("player")

const recordingTime = 3000;
const timerStep = 40
let interval;
let time = 0;

micBtn.onclick = () => {
  micBtn.disabled = true;
  micBtn.classList.toggle("recording");

  time = 0;
  updateTimer();

  interval = setInterval(() => {
    time += timerStep;
    updateTimer();

    if (time === recordingTime) { 
      clearInterval(interval);
      micBtn.classList.toggle("recording");
      micBtn.disabled = false;

      playerDialog.showModal()
      player.hidden = false
    }
  }, timerStep);
};

function updateTimer() {
  const sec = Math.floor(time / 1000);
  const milliSec = Math.floor(time % 1000 / 10);
  timer.innerText = `${String(sec).padStart(2, "0")}.${String(milliSec).padStart(2, "0")}`;
}
