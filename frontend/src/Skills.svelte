<script>
  import { onMount } from 'svelte';
  import { t as _ } from './lib/i18n.js';

  let skills = [];
  let search = '';
  let toolFilter = '';
  let category = '';
  let detail = null;

  async function load() {
    try {
      skills = await window.go.main.App.GetSkills(toolFilter, category, search);
    } catch(e) {}
  }

  function showDetail(s) { detail = s; }
  function closeDetail() { detail = null; }
  function onSearch() { load(); }

  onMount(() => { load(); });
</script>

<div>
  <h2 style="margin-bottom:16px;font-size:20px">{_('skills.title')}</h2>

  <div class="toolbar">
    <input type="text" placeholder={_('skills.search')} bind:value={search} on:input={onSearch} />
    <select bind:value={toolFilter} on:change={onSearch}>
      <option value="">{_('skills.allTools')}</option>
      <option value="claude">Claude Code</option>
      <option value="opencode">OpenCode</option>
      <option value="trae">Trae</option>
    </select>
    <select bind:value={category} on:change={onSearch}>
      <option value="">{_('skills.allCategories')}</option>
      <option value="skills">Skills</option>
      <option value="rules">Rules</option>
      <option value="agents">Agents</option>
      <option value="commands">Commands</option>
      <option value="memories">Memories</option>
    </select>
  </div>

  {#if detail}
    <button class="back" on:click={closeDetail}>{_('skills.back')}</button>
    <div class="detail-card">
      <h3>{detail.name}</h3>
      <p class="desc">{detail.summary}</p>
      <div class="tags">{detail.tool} · {detail.category} · {detail.size} {_('skills.bytes')}</div>
    </div>
  {:else}
    <div class="skill-grid">
      {#each skills as s}
        <div class="skill-card" on:click={() => showDetail(s)} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && showDetail(s)}>
          <div class="name">{s.name}</div>
          <div class="meta">{s.tool} · {s.category}</div>
          <div class="size">{s.size} {_('skills.bytes')}</div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .toolbar { display:flex; gap:8px; margin-bottom:16px; }
  .toolbar input, .toolbar select { background:#1a1a28; border:1px solid #2a2a3a; color:#ccc; padding:8px 12px; border-radius:6px; font-size:13px; }
  .toolbar input:focus, .toolbar select:focus { border-color:#7c5cfc; outline:none; }
  .skill-grid { display:grid; grid-template-columns:repeat(auto-fill, minmax(260px, 1fr)); gap:12px; }
  .skill-card { background:#1a1a28; border:1px solid #2a2a3a; border-radius:8px; padding:14px; cursor:pointer; transition:.15s; }
  .skill-card:hover { border-color:#7c5cfc; background:#1e1e2e; }
  .skill-card .name { font-size:15px; font-weight:600; margin-bottom:4px; }
  .skill-card .meta { font-size:11px; color:#555; }
  .skill-card .size { font-size:11px; color:#444; margin-top:4px; }
  .back { background:none; border:none; color:#888; cursor:pointer; font-size:13px; margin-bottom:12px; }
  .back:hover { color:#ccc; }
  .detail-card { background:#1a1a28; border:1px solid #7c5cfc; border-radius:8px; padding:20px; }
  .detail-card h3 { color:#7c5cfc; margin-bottom:8px; }
  .desc { color:#aaa; line-height:1.6; margin-bottom:12px; }
  .tags { color:#555; font-size:12px; }
</style>