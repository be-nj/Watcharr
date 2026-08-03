<script lang="ts">
	import Error from "@/lib/Error.svelte";
	import Spinner from "@/lib/Spinner.svelte";
	import CreateSourceModal from "@/lib/source/CreateSourceModal.svelte";
	import CinemaDetailsModal from "@/lib/source/CinemaDetailsModal.svelte";
	import { req } from "@/lib/util/api";
	import { notify } from "@/lib/util/notify";
	import type { WatchSource } from "@/types";

	let sources = $state<WatchSource[]>([]);
	let createModalOpen = $state(false);
	let sourceToEdit: WatchSource | undefined = $state(undefined);
	let cinemaToEdit: WatchSource | undefined = $state(undefined);
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

	async function deleteSource(source: WatchSource) {
		const nid = notify({ text: "Deleting Source", type: "loading" });
		try {
			await req.delete(`/source/${source.id}`);
			sources = sources.filter((s) => s.id !== source.id);
			notify({ id: nid, text: "Source Deleted!", type: "success" });
		} catch (err) {
			console.error("deleteSource: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
		}
	}

	function onModalClose(updated?: WatchSource) {
		createModalOpen = false;
		sourceToEdit = undefined;
		if (updated) {
			getSourcesPromise = getSources();
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
			<button onclick={() => (createModalOpen = true)}>Add Source</button>
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
						<div class="source">
							<span class="name">
								{source.name}
								{#if source.type === "CINEMA" && source.cinema?.city}
									<span class="city">{source.cinema.city}</span>
								{/if}
								{#if source.type === "CINEMA" && source.cinema?.ratingOverall}
									<span class="rating">
										{source.cinema.ratingOverall}/10
									</span>
								{/if}
							</span>
							<span class="actions">
								{#if source.type === "CINEMA"}
									<button class="plain" onclick={() => (cinemaToEdit = source)}>
										Details
									</button>
								{/if}
								<button class="plain" onclick={() => (sourceToEdit = source)}>
									Edit
								</button>
								<button class="plain" onclick={() => deleteSource(source)}>
									Delete
								</button>
							</span>
						</div>
					{/each}
				</div>
			{/each}
		{:catch err}
			<Error error={err} pretty="Failed to load sources!" />
		{/await}
	</div>
</div>

{#if createModalOpen || sourceToEdit}
	<CreateSourceModal onClose={onModalClose} existingSource={sourceToEdit} />
{/if}

{#if cinemaToEdit}
	<CinemaDetailsModal
		source={cinemaToEdit}
		onClose={() => (cinemaToEdit = undefined)}
	/>
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

			.name {
				display: flex;
				align-items: center;
				gap: 10px;

				.city,
				.rating {
					font-size: 13px;
					opacity: 0.7;
				}
			}

			.actions {
				display: flex;
				gap: 6px;

				button {
					font-size: 13px;
					width: max-content;
				}
			}
		}
	}
</style>
