<script lang="ts">
	import { resolve } from "$app/paths";
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import CreateSourceModal from "@/lib/source/CreateSourceModal.svelte";
	import { req } from "@/lib/util/api";
	import type { WatchSource } from "@/types";
	import { goto } from "$app/navigation";

	let sources = $state<WatchSource[]>([]);
	let createModalOpen = $state(false);
	let getSourcesPromise = $state(getSources());

	const typeNames: { [key: string]: string } = {
		CINEMA: "Cinemas",
		STREAMING: "Streaming",
		TV: "TV",
		SELFHOSTED: "Selfhosted",
		DISC: "Disc",
		OTHER: "Other",
	};

	let groupedSources = $derived(
		Object.keys(typeNames)
			.map((t) => ({
				type: t,
				name: typeNames[t],
				sources: sources.filter((s) => s.type === t),
			}))
			.filter((g) => g.sources.length > 0),
	);

	async function getSources() {
		sources = await req.get<WatchSource[]>("/source");
	}

	function onModalClose(created?: WatchSource) {
		createModalOpen = false;
		if (created) {
			// Neu angelegte Quelle direkt auf ihrer Seite oeffnen.
			goto(resolve(`/sources/${created.id}`));
		}
	}
</script>

<svelte:head>
	<title>Watch Sources</title>
</svelte:head>

<div class="content">
	<div class="inner">
		<div class="header">
			<h2>Watch Sources</h2>
			<p>
				Where/how you watch things. Add your cinemas and services here, then
				attach them to your watches from the activity list.
			</p>
			<div class="header-btns">
				<button onclick={() => (createModalOpen = true)}>Add Source</button>
				<a href={resolve("/sources/map")}>
					<button>Cinema Map</button>
				</a>
			</div>
		</div>
		{#await getSourcesPromise}
			<Spinner />
		{:then}
			{#if sources.length === 0}
				<p class="empty">No sources yet. Add your first one!</p>
			{/if}
			{#each groupedSources as group (group.type)}
				<h3>{group.name}</h3>
				<div class="sources">
					{#each group.sources as source (source.id)}
						<a href={resolve(`/sources/${source.id}`)} class="source">
							<span class="name">{source.name}</span>
							<span class="meta">
								{#if source.type === "CINEMA" && source.cinema?.city}
									<span>{source.cinema.city}</span>
								{/if}
								{#if source.type === "CINEMA" && source.cinema?.ratingOverall}
									<span>{source.cinema.ratingOverall}/10</span>
								{/if}
							</span>
						</a>
					{/each}
				</div>
			{/each}
		{:catch err}
			<Error error={err} pretty="Failed to load sources!" />
		{/await}
	</div>
</div>

{#if createModalOpen}
	<CreateSourceModal onClose={onModalClose} />
{/if}

<style lang="scss">
	.content {
		display: flex;
		width: 100%;
		justify-content: center;
		padding: 0 30px 30px 30px;

		.inner {
			display: flex;
			flex-flow: column;
			min-width: 400px;
			max-width: 700px;
			width: 100%;

			@media screen and (max-width: 420px) {
				min-width: 100%;
			}
		}
	}

	.header {
		display: flex;
		flex-flow: column;
		gap: 8px;
		margin-bottom: 20px;

		button {
			width: max-content;
		}

		.header-btns {
			display: flex;
			gap: 8px;
		}
	}

	h3 {
		margin: 15px 0 8px 0;
	}

	.empty {
		font-style: italic;
	}

	.sources {
		display: flex;
		flex-flow: column;
		gap: 6px;

		.source {
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding: 8px 12px;
			border-radius: 8px;
			background-color: $accent-color;
			text-decoration: none;

			.meta {
				display: flex;
				gap: 10px;
				font-size: 13px;
				opacity: 0.7;
			}
		}
	}
</style>
