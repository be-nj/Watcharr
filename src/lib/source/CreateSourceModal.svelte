<script lang="ts">
	import Modal from "../Modal.svelte";
	import Setting from "../settings/Setting.svelte";
	import SettingsList from "../settings/SettingsList.svelte";
	import DropDown from "../DropDown.svelte";
	import { notify } from "../util/notify";
	import { req } from "../util/api";
	import type {
		WatchSource,
		WatchSourceAddRequest,
		WatchSourceType,
	} from "@/types";
	import { onMount } from "svelte";

	interface Props {
		onClose: (updated?: WatchSource) => void;
		// Passing an existing source will enable 'Edit Source' mode.
		existingSource?: WatchSource | undefined;
	}

	let { onClose, existingSource = undefined }: Props = $props();

	const sourceTypes = [
		{ id: "CINEMA", value: "Cinema" },
		{ id: "STREAMING", value: "Streaming" },
		{ id: "TV", value: "TV" },
		{ id: "SELFHOSTED", value: "Selfhosted" },
		{ id: "DISC", value: "Disc" },
		{ id: "OTHER", value: "Other" },
	];

	let sourceName = $state("");
	let sourceType: string | number | undefined = $state(undefined);
	let error = $state("");
	let submitDisabled = $state(false);
	let modalTitle = $state("Create A Watch Source");
	let modalDesc = $state("Create a new watch source");
	let submitBtnText = $state("Create Source");

	async function addSource() {
		console.debug("addSource:", sourceName, sourceType);
		if (!sourceName) {
			error = "Source must have a name!";
			return;
		}
		if (!sourceType) {
			error = "Source must have a type!";
			return;
		}
		const nid = notify({ text: "Creating Source", type: "loading" });
		try {
			const resp = await req.post<WatchSource>("/source", {
				name: sourceName,
				type: sourceType as WatchSourceType,
			} as WatchSourceAddRequest);
			console.log("addSource: Source was created", resp);
			notify({ id: nid, text: "Source Created!", type: "success" });
			onClose(resp);
		} catch (err) {
			console.error("addSource: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
			error = "Failed!";
		}
	}

	async function updateSource() {
		console.debug("updateSource:", existingSource, sourceName, sourceType);
		if (!sourceName) {
			error = "Source must have a name!";
			return;
		}
		const nid = notify({ text: "Modifying Source", type: "loading" });
		try {
			await req.put(`/source/${existingSource!.id}`, {
				name: sourceName,
				type: sourceType as WatchSourceType,
			} as WatchSourceAddRequest);
			existingSource!.name = sourceName;
			existingSource!.type = sourceType as WatchSourceType;
			notify({ id: nid, text: "Source Modified!", type: "success" });
			onClose(existingSource);
		} catch (err) {
			console.error("updateSource: Failed!", err);
			notify({ id: nid, text: "Failed!", type: "error", time: 1 });
			error = "Failed!";
		}
	}

	async function submitClicked() {
		submitDisabled = true;
		try {
			if (existingSource) {
				await updateSource();
			} else {
				await addSource();
			}
		} catch (err) {
			console.log("CreateSourceModal: Submit failed!", err);
		}
		submitDisabled = false;
	}

	onMount(() => {
		if (existingSource) {
			console.log(
				"CreateSourceModal: Entering edit mode for source:",
				existingSource,
			);
			modalTitle = "Edit Watch Source";
			modalDesc = "Edit an existing watch source";
			submitBtnText = "Edit Source";
			sourceName = existingSource.name;
			sourceType = existingSource.type;
		}
	});
</script>

<div class="wrap">
	<Modal
		title={modalTitle}
		desc={modalDesc}
		maxWidth="500px"
		onClose={() => onClose()}
		{error}
	>
		<SettingsList>
			<Setting title="Name" desc="What should we call this source?">
				<input
					type="text"
					name="name"
					placeholder="Name"
					bind:value={sourceName}
				/>
			</Setting>
			<Setting
				title="Type"
				desc="What kind of source is this? Cinemas can have extra details."
			>
				<DropDown
					options={sourceTypes}
					isDropDownItem={true}
					bind:active={sourceType}
					placeholder="Type"
				/>
			</Setting>
			<button
				class="add-source-btn"
				onclick={() => submitClicked()}
				disabled={submitDisabled}
			>
				{submitBtnText}
			</button>
		</SettingsList>
	</Modal>
</div>

<style lang="scss">
	.add-source-btn {
		width: max-content;
		margin-left: auto;
	}

	.wrap {
		color: $text-color;
	}
</style>
