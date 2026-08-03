<script lang="ts">
	import Notifications from "@/lib/Notifications.svelte";
	import { onMount } from "svelte";
	import { pwaInfo } from "virtual:pwa-info";

	interface Props {
		children?: import("svelte").Snippet;
	}

	let { children }: Props = $props();

	function resetTooltipPos() {
		const t = document.getElementById("tooltip");
		if (t) {
			t.style.top = "0";
			t.style.left = "0";
		}
	}

	onMount(() => {
		window.addEventListener("resize", resetTooltipPos);

		return () => {
			window.removeEventListener("resize", resetTooltipPos);
		};
	});
</script>

<svelte:head>
	{#if pwaInfo?.webManifest}
		<!-- Absolute href: the generated link tag is relative, which
		     resolves to eg /tv/manifest.webmanifest on sub pages and
		     serves html instead of the manifest (upstream #823). -->
		<link rel="manifest" href="/manifest.webmanifest" />
	{/if}
</svelte:head>

<div id="tooltip"></div>

<Notifications />

{@render children?.()}

<style lang="scss">
	@use "../styles/norm.scss";
</style>
