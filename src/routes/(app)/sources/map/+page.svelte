<script lang="ts">
	import { resolve } from "$app/paths";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import { req } from "@/lib/util/api";
	import type { WatchSource } from "@/types";
	import { onDestroy, onMount } from "svelte";
	import "leaflet/dist/leaflet.css";

	let mapElement: HTMLDivElement | undefined = $state();
	let map: import("leaflet").Map | undefined;
	let cinemasWithCoords = $state(0);
	let loadPromise = $state<Promise<void> | undefined>(undefined);

	async function load() {
		// Leaflet greift auf window zu, deshalb erst im Browser importieren.
		const L = (await import("leaflet")).default;
		const sources = await req.get<WatchSource[]>("/source");
		const cinemas = sources.filter(
			(s) =>
				s.type === "CINEMA" &&
				s.cinema?.lat !== undefined &&
				s.cinema?.lat !== null &&
				s.cinema?.lon !== undefined &&
				s.cinema?.lon !== null,
		);
		cinemasWithCoords = cinemas.length;
		if (!mapElement) {
			return;
		}
		map = L.map(mapElement);
		L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
			maxZoom: 19,
			attribution:
				'&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
		}).addTo(map);
		// Standard-Marker-Icons aus dem Paket beziehen (Bundler-Pfade).
		const icon = L.icon({
			iconUrl: (await import("leaflet/dist/images/marker-icon.png")).default,
			iconRetinaUrl: (await import("leaflet/dist/images/marker-icon-2x.png"))
				.default,
			shadowUrl: (await import("leaflet/dist/images/marker-shadow.png"))
				.default,
			iconSize: [25, 41],
			iconAnchor: [12, 41],
			popupAnchor: [1, -34],
			shadowSize: [41, 41],
		});
		const bounds = L.latLngBounds([]);
		for (const c of cinemas) {
			const pos: [number, number] = [c.cinema!.lat!, c.cinema!.lon!];
			bounds.extend(pos);
			const rating = c.ratingAverage
				? `<br/>&Oslash; ${c.ratingAverage.toFixed(1)}/10 (${c.ratingCount} rated ${c.ratingCount === 1 ? "visit" : "visits"})`
				: "";
			const snacks = "";
			L.marker(pos, { icon })
				.addTo(map)
				.bindPopup(
					`<b><a href="${resolve(`/sources/${c.id}`)}">${c.name}</a></b>` +
						`${c.cinema?.city ? `<br/>${c.cinema.city}` : ""}${rating}${snacks}`,
				);
		}
		if (cinemas.length > 0) {
			map.fitBounds(bounds, { padding: [40, 40] });
		} else {
			map.setView([51.16, 10.45], 6); // Deutschland
		}
	}

	onMount(() => {
		loadPromise = load();
	});

	onDestroy(() => {
		map?.remove();
	});
</script>

<svelte:head>
	<title>Cinema Map</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<a href={resolve("/sources")} class="back">← All Sources</a>
		<h2>Cinema Map</h2>
		{#if loadPromise}
			{#await loadPromise}
				<Spinner />
			{:then}
				{#if cinemasWithCoords === 0}
					<p class="empty">
						No cinemas with coordinates yet. Add them on the source pages.
					</p>
				{/if}
			{:catch err}
				<Error error={err} pretty="Failed to load map!" />
			{/await}
		{/if}
		<div class="map" bind:this={mapElement}></div>
	</div>
</div>

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 30px 30px;

		.inner {
			display: flex;
			flex-flow: column;
			width: 100%;
			max-width: 1100px;
		}
	}

	.back {
		margin-bottom: 10px;
		width: max-content;
	}

	h2 {
		margin-bottom: 15px;
	}

	.empty {
		font-style: italic;
		margin-bottom: 10px;
	}

	.map {
		width: 100%;
		height: 70vh;
		border-radius: 10px;
	}
</style>
