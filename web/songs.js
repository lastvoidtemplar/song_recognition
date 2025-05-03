import { ApiHandler } from "./api_handler.js";

const songsHeader = document.getElementById("songs-header")
const openDialogButton = document.getElementById("open-dialog")
const songsTable = document.getElementById("songs");
const songsTableBody = document.getElementById("songs-body");
const pager = document.getElementById("pager");
const spinner = document.getElementById("spinner");
const errorDisplay = document.getElementById("error-display");
const songsDialog = document.getElementById("songs-dialog")
const addSongBtn = document.getElementById("add-song")

const url = new URL("/songs", window.location.origin);
const params = {
  page: 1,
  limit: 14,
};
Object.entries(params).forEach(([key, value]) => {
  url.searchParams.append(key, value);
});

const songsHandler = new ApiHandler(url.toString());

songsHandler.onLoading(() => {
  spinner.hidden = false;
  console.log("loading");
});

songsHandler.onSuccess((data) => {
  spinner.hidden = true;
  songsTable.hidden = false;
  pager.hidden = false;

  renderSongsTable(data.songs, 1);

  const pageCount = Math.ceil(data.total / data.limit);

  renderPager(1, pageCount);
  renderSongHeaders()
});

songsHandler.onError((statusCode, err) => {
  spinner.hidden = true;
  errorDisplay.innerText = `Status code - ${statusCode}, error - ${err.error}`;
});

songsHandler.onFail((err) => {
  spinner.hidden = true;
  errorDisplay.innerText = "Couldn`t fetch song database!";
});

songsHandler.initiateFetch();

function renderSongsTable(songs, num) {
  songsTableBody.innerHTML = "";

  for (const song of songs) {
    const node = document.createElement("tr");
    node.innerHTML = `<td>${num}.</td><td>${song.song_title}</td><td>${song.song_url}</td>`;
    songsTableBody.append(node);
    num++;
  }
}

function renderPager(page, pageCount) {
  console.log(page, pageCount);

  pager.innerHTML = "";

  const min = Math.min(pageCount, 3);
  for (let num = 1; num <= min; num++) {
    const node = document.createElement("button");
    node.innerText = num;
    if (num === page) {
      node.classList.add("active");
      node.disabled = true;
    }
    node.onclick = createOnClick(num);
    pager.append(node);
  }

  if (page > 5) {
    const node = document.createElement("p");
    node.innerText = "...";
    pager.append(node);
  }

  for (let ind = -1; ind <= 1; ind++) {
    console.log(page + ind);
    if (page + ind > 3 && page + ind <= pageCount) {
      const node = document.createElement("button");
      node.innerText = page + ind;
      if (ind === 0) {
        node.classList.add("active");
        node.disabled = true;
      }
      node.onclick = createOnClick(page + ind);
      pager.append(node);
    }
  }

  if (pageCount - page > 4) {
    const node = document.createElement("p");
    node.innerText = "...";
    pager.append(node);
  }

  const max = Math.max(page + 2, pageCount - 2);
  for (let num = max; num <= pageCount; num++) {
    const node = document.createElement("button");
    node.innerText = num;
    if (num === page) {
      node.classList.add("active");
      node.disabled = true;
    }
    node.onclick = createOnClick(num);
    pager.append(node);
  }
}

function createOnClick(num) {
  return () => {
    pager.childNodes.forEach((node) => {
      node.disabled = true;
    });
    const params = {
      page: num,
      limit: 14,
    };
    const url = new URL("/songs", window.location.origin);
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.append(key, value);
    });

    const pageHandler = new ApiHandler(url.toString());

    pageHandler.onLoading(() => {
      spinner.hidden = false;
      songsTable.style.opacity = 0.5;
      console.log("loading");
    });

    pageHandler.onSuccess((data) => {
      spinner.hidden = true;
      songsTable.style.opacity = 1;

      const num = (data.page - 1) * data.limit + 1;
      renderSongsTable(data.songs, num);

      const pageCount = Math.ceil(data.total / data.limit);

      renderPager(data.page, pageCount);
    });
    pageHandler.initiateFetch();
  };
}

function renderSongHeaders(){
  songsHeader.childNodes.forEach(el=>el.hidden = false)
  openDialogButton.onclick = ()=>{
    console.log("open");
    
    songsDialog.showModal()
  }
}
