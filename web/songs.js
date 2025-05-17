import { ApiHandler } from "./api_handler.js";
import {API_URL} from "./api_url.js"

const songsHeader = document.getElementById("songs-header");
const openDialogButton = document.getElementById("open-dialog");
const songsTable = document.getElementById("songs");
const songsTableBody = document.getElementById("songs-body");
const pager = document.getElementById("pager");
const spinner = document.getElementById("spinner");
const errorDisplay = document.getElementById("error-display");
const songsDialog = document.getElementById("songs-dialog");
const songUrlInput = document.getElementById("song-url");
const addSongBtn = document.getElementById("add-song");
const errorDialog = document.getElementById("error-dialog");

const apiUrl = API_URL;
const limit = 14;
const url = new URL("/songs", apiUrl);

const params = {
  page: 1,
  limit: limit,
};

Object.entries(params).forEach(([key, value]) => {
  url.searchParams.append(key, value);
});

const songsHandler = new ApiHandler(url.toString());

songsHandler.onLoading(() => {
  spinner.hidden = false;
});

songsHandler.onSuccess((data) => {
  if (!data) {
    spinner.hidden = true;
    errorDisplay.innerText = "Couldn`t connect to the server";
    return;
  }

  spinner.hidden = true;
  songsTable.hidden = false;
  pager.hidden = false;

  renderSongsTable(data.songs, data.limit);

  const pageCount = Math.ceil(data.total / data.limit);

  renderPager(1, pageCount);
  renderSongHeaders();
});

songsHandler.onError((statusCode, err) => {
  spinner.hidden = true;
  errorDisplay.innerText = `Status code - ${statusCode}, error - ${err.error}`;
});

songsHandler.onFail((err) => {
  spinner.hidden = true;
  errorDisplay.innerText = "Couldn`t connect to the server";
});

songsHandler.initiateFetch();

function renderSongsTable(songs, limit) {
  songsTableBody.innerHTML = "";

  for (let i = 0; i < limit; i++) {
    const node = document.createElement("tr");
    if (i < songs.length) {
      const song = songs[i];
      node.innerHTML = `<td>${song.song_id}.</td><td>${song.song_title}</td><td>${song.song_url}</td>`;
    } else {
      node.innerHTML = "<td>&nbsp;</td><td>&nbsp;</td><td>&nbsp;</td>";
    }
    songsTableBody.append(node);
  }
}

function renderPager(page, pageCount) {
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
      limit: limit,
    };
    const url = new URL("/songs", apiUrl);
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.append(key, value);
    });

    const pageHandler = new ApiHandler(url.toString());

    pageHandler.onLoading(() => {
      spinner.hidden = false;
      songsTable.style.opacity = 0.5;
    });

    pageHandler.onSuccess((data) => {
      if (!data) {
        spinner.hidden = true;
        errorDisplay.innerText = "Couldn`t connect to the server";
        return;
      }

      spinner.hidden = true;
      songsTable.style.opacity = 1;

      const num = (data.page - 1) * data.limit + 1;
      renderSongsTable(data.songs, data.limit);

      const pageCount = Math.ceil(data.total / data.limit);

      renderPager(data.page, pageCount);
    });
    pageHandler.initiateFetch();
  };
}

function renderSongHeaders() {
  songsHeader.childNodes.forEach((el) => (el.hidden = false));
  openDialogButton.onclick = () => {
    errorDialog.innerText = "";
    errorDialog.hidden = true;
    songUrlInput.value = "";
    songsDialog.showModal();
  };
}

songsDialog.onclick = (e) => {
  if (e.target === songsDialog) {
    errorDialog.innerText = "";
    errorDialog.hidden = true;
    songsDialog.close();
  }
};

addSongBtn.onclick = () => {
  const songUrl = songUrlInput.value;

  const addSongHandler = new ApiHandler(
    "/songs",
    "post",
    JSON.stringify({
      song_url: songUrl,
    })
  );

  addSongHandler.onSuccess((data) => {
    console.log("Uploaded");
    songsDialog.close();
  });

  addSongHandler.onError((_, err) => {
    errorDialog.innerText = err.error;
    errorDialog.hidden = false;
  });

  addSongHandler.onFail((err) => {
    errorDialog.innerText = "Couldn`t connect to the server";
    errorDialog.hidden = false;
  });

  addSongHandler.initiateFetch();
};
