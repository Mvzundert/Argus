package web

import (
	"argus/config"
	"argus/services"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// IndexHandler serves the main HTML page.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Now Playing</title>
			<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
			<style type="text/tailwindcss">
				@tailwind base;
				@tailwind components;
				@tailwind utilities;

				@layer base {
					body {
						font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji";
					}
				}
			</style>
			<style>
				/* Your custom CSS for animations */
				.animate {
					animation: move 9.48s linear 0s infinite alternate;
				}
				.animated-title {
					overflow: hidden;
					width: 100%;
					white-space: nowrap;
				}
				.animated-title > * {
					display: inline-block;
					position: relative;
				}
				.animated-title > *.min {
					min-width: 100%;
				}
				@keyframes move {
					0%, 25% {
						transform: translate(0);
						left: 0%;
					}
					75%, to {
						transform: translate(-100%);
						left: 100%;
					}
				}
			</style>
		</head>
		<body class="bg-transparent text-white flex items-center justify-center min-h-screen">
			<div id="spotify-widget" class="flex flex-col p-6 rounded-xl shadow-lg w-full max-w-lg" style="background-color: rgba(31, 41, 55, 0.7);">
				<p id="artist-name" class="text-xl text-gray-200 font-semibold mb-1">Loading Artist...</p>
				<div id="title-container" class="relative z-50 overflow-clip whitespace-nowrap text-lg font-bold animated-title mb-4">
					<h2 id="song-title" class="text-3xl font-bold text-white"><span>Loading Song...</span></h2>
				</div>
				<div class="w-full h-2 bg-gray-500 rounded-full">
					<div id="progress-bar" class="h-full bg-red-600 rounded-full transition-all duration-100" style="width: 0%;"></div>
				</div>
			</div>
		<script>
			function updateNowPlaying() {
				fetch('/now-playing')
					.then(response => response.json())
					.then(data => {
						const widget = document.getElementById('spotify-widget');
						const songTitleEl = document.getElementById('song-title');
						const artistNameEl = document.getElementById('artist-name');
						const progressBarEl = document.getElementById('progress-bar');
						const titleContainer = document.getElementById('title-container');

						if (data && data.is_playing) {
							widget.style.display = 'flex';
							artistNameEl.textContent = data.item.artists[0].name;
							songTitleEl.textContent = data.item.name;

							// Reset animation state
							songTitleEl.classList.remove('animate');
							songTitleEl.style.transform = 'translate(0)';
							songTitleEl.style.left = '0%';

							if (songTitleEl.scrollWidth > titleContainer.clientWidth) {
								songTitleEl.classList.add('animate');
							}

							const progressMs = data.progress_ms;
							const durationMs = data.item.duration_ms;

							if (durationMs > 0) {
								const progressPercentage = (progressMs / durationMs) * 100;
								progressBarEl.style.width = progressPercentage + '%';
							} else {
								progressBarEl.style.width = '0%';
							}
						} else {
							widget.style.display = 'none';
						}
					})
					.catch(error => console.error('Error fetching data:', error));
			}

			updateNowPlaying();
			setInterval(updateNowPlaying, 1000);
		</script>
		</body>
		</html>
	`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

// NowPlayingHandler handles requests for the current song.
func NowPlayingHandler(nowPlayingSvc services.NowPlayingService, w http.ResponseWriter, r *http.Request) {
	data, err := nowPlayingSvc.GetNowPlayingInfo()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Error getting now playing info: %v", err)
		json.NewEncoder(w).Encode(services.NowPlayingData{IsPlaying: false})
		return
	}

	json.NewEncoder(w).Encode(data)
}

// StartServer starts the web server.
func StartServer(cfg config.Config) error {
	nowPlayingSvc := services.NewNowPlayingService()

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/now-playing", func(w http.ResponseWriter, r *http.Request) {
		NowPlayingHandler(nowPlayingSvc, w, r)
	})

	if cfg.ShowLogs {
		log.Printf("Starting server on :%s", cfg.Port)
	}
	// Return the error from ListenAndServe instead of calling log.Fatalf.
	return http.ListenAndServe(":"+cfg.Port, nil)
}
