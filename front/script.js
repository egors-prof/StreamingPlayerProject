const showMore=document.querySelector("#show-more");
const showMoreBtn=document.querySelector("#show-more-button");
const musicTemplate=document.querySelector("[data-music-template]");
let musicContainer=document.querySelector("[data-music-container]");
const check=document.querySelector("#check");

let songsQuantity=5;
let songsTotal=[];
let musicObj=[]
const username = document.querySelector("#user-name");
const searchInput=document.querySelector("#searchInput");
searchInput.addEventListener("input",async (e)=>{
    showMoreBtn.classList.add("hide");
    const response=await fetch(`http://localhost:8888/search?q=${searchInput.value}`);
    const search = await response.json();
    musicContainer.innerHTML=""
    showInResult(search);






})

let source = null;
let audioChunks = [];
let totalSize = 0;
let receivedSize = 0;
let audioContextReady = false;




function showInResult(songs){
    for (let song in songs){
        console.log(songs[song]);
        console.log(songs[song].photo_path);
        const musicCard=musicTemplate.content.cloneNode(true).children[0]
        console.log(musicCard)
        const songName =musicCard.querySelector("[data-name]")
        const songAlbum=musicCard.querySelector("[data-album]")
        const songArtist=musicCard.querySelector("[data-artist]")
        const songDuration=musicCard.querySelector("[data-duration]")
        const playImg=musicCard.children[0].children[0].children[0]
        const playButton=musicCard.querySelector("[data-play-button]")


        songName.textContent=songs[song].title
        songAlbum.textContent=songs[song].album_title;
        songDuration.textContent=songs[song].duration;
        songArtist.textContent=songs[song].pseudonym;
        playImg.style.backgroundImage=`url(${songs[song].photo_path})`
        playImg.classList.add("bg-contain")

        playButton.addEventListener("click", async () => {
            const songTitle = songs[song].title;
            let svgPath = playButton.children[0].children[0];

            const isCurrentSong = currentPlayingSong === songTitle;

            // Pause current song
            if (isCurrentSong && isPlaying) {
                if (source) {
                    playbackPosition = ctx.currentTime - startTime;
                    source.stop();
                    source = null;
                    isPlaying = false;
                    svgPath.setAttribute('d', playPath);
                }
                return;
            }

            // Resume paused song
            if (isCurrentSong && !isPlaying) {
                // Clear current chunks if any
                chunks.length = 0;

                // Request the song from the server with the current position
                ws.send("play:" + songTitle + ":" + playbackPosition);
                svgPath.setAttribute('d', pausePath);
                isPlaying = true;
                return;
            }

            // New song or different song
            if (currentPlayingSong && currentPlayingSong !== songTitle) {
                // Stop current audio
                if (source) {
                    source.stop();
                    source = null;
                }
                chunks.length = 0;
                playbackPosition = 0;

                // Reset previous song's button
                for (let m in musicObj) {
                    if (musicObj[m].title.textContent === currentPlayingSong) {
                        musicObj[m].playButton.children[0].children[0].setAttribute('d', playPath);
                        musicObj[m].isPlaying = false;
                        break;
                    }
                }
            }

            // Play new song
            svgPath.setAttribute('d', pausePath);
            currentPlayingSong = songTitle;
            isPlaying = true;
            playbackPosition = 0;

            // Update musicObj state
            for (let m in musicObj) {
                if (musicObj[m].title.textContent === currentPlayingSong) {
                    musicObj[m].isPlaying = true;
                    continue;
                }
                musicObj[m].isPlaying = false;
                musicObj[m].playButton.children[0].children[0].setAttribute('d', playPath);
            }

            // Request new song
            ws.send("play:" + songTitle);
        });




        musicContainer.appendChild(musicCard)
        musicObj.push({
            title:songName,
            album:songAlbum,
            photo_path:song.photo_path,
            duration:songDuration,
            artist:songArtist,
            playButton:playButton,
            card:musicCard,
            isPlaying:false,
        })
    }
}

ctx = new AudioContext();
const chunks=[]



function cutOut(str){
    let alphaPos
    for (let i = 0; i < str.length; i++) {
        if (str[i]==='@'){
            alphaPos = i
        }
    }
    return str.slice(0, alphaPos);
}
//audio

let ws = new WebSocket("ws://localhost:9999/stream/ws");
function SetupWebsocket(){
    ws.onopen=async ()=>{
        console.log("ws open");

        ws.send(JSON.stringify(
            {
                "data_type":'auth',
                "access_token":localStorage.getItem("refresh_token"),
            }
        ))


    }
    ws.onmessage = async (event) => {
        if (typeof event.data === 'string') {
            if (event.data === "complete") {
                console.log("Audio download complete");
                await playStream()
            }else if (event.data === "auth success") {
                console.log("authentication success");
            }else{
                const response=JSON.parse(event.data);
                console.log(response);
                const greet=document.getElementById("greeting");
                user={
                    userId:response.user_id,
                    username: response.username,
                    role:response.role,


                }
                greet.innerHTML='Welcome back '+cutOut(user.username);
                username.textContent=cutOut(user.username);
                console.log("Greet message",user);

            }
        } else if (event.data instanceof Blob) {
            chunks.push(event.data);


        }
    }

    ws.onclose = () => console.log('Disconnected');
    ws.onerror = (err) => console.log('WebSocket error:', err);
}

SetupWebsocket();



//initial fetch
getFirstSongs(5,0)

showMore.addEventListener("click",async ()=>{

    showMore.classList.add("mb-10");


    let songs=await getFirstSongs(10,songsQuantity)
    songsTotal.push(songs);


    songsQuantity+=10;
})



let data=[]

let previous

let playPath='M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z'
let pausePath='M9 6h2v12H9zM13 6h2v12h-2z'

async function getFirstSongs(quant,off){

    const url=`http://localhost:8888?quant=${quant}&off=${off}`;
    showMoreBtn.classList.remove("hide");
    data=await fetch(url)
    let songs = await data.json()
    if(songs.length === 0){
        showMoreBtn.disabled=true;
        showMoreBtn.style.color = "gray";
        check.style.color = "gray";
        showMoreBtn.classList.remove("hover:bg-slate-800");



    }

    showInResult(songs);



    return songs



}




let isPlaying = false;
let currentPlayingSong = null;
let playbackPosition = 0;
let startTime = 0;
let pausedTime = 0;

// Enhanced playStream function with pause/resume support
async function playStream(startFrom = 0) {
    if (chunks.length === 0) return;

    const blob = new Blob(chunks);
    const buffer = await blob.arrayBuffer();
    const audioBuffer = await ctx.decodeAudioData(buffer);

    source = ctx.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(ctx.destination);

    // Start from the specified position
    if (startFrom > 0) {
        source.start(0, startFrom);
    } else {
        source.start();
    }

    startTime = ctx.currentTime - startFrom;

    source.onended = () => {
        source = null;
        chunks.length = 0;
        isPlaying = false;
        currentPlayingSong = null;
        playbackPosition = 0;
        pausedTime = 0;

        // Reset all play buttons
        for (let m in musicObj) {
            musicObj[m].playButton.children[0].children[0].setAttribute('d', playPath);
            musicObj[m].isPlaying = false;
        }
    };
}






