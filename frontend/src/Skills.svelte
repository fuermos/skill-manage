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
      <h3>{detail.display_name || detail.name}</h3>
      <div class="meta-bar">
        <span class="badge">{detail.tool}</span>
        <span class="badge">{detail.category}</span>
        <span class="badge">{detail.size} {_('skills.bytes')}</span>
      </div>
      {#if detail.description}
        <p class="desc">{detail.description}</p>
      {:else if detail.summary}
        <p class="desc">{detail.summary}</p>
      {/if}
      {#if detail.tags && detail.tags.length}
        <div class="tags-bar">
          {#each detail.tags as tag}
            <span class="tag">{tag}</span>
          {/each}
        </div>
      {/if}
    </div>
  {:else}
    <div class="skill-grid">
      {#each skills as s}
        <div class="skill-card" on:click={() => showDetail(s)} role="button" tabindex="0" on:keydown={(e) => e.key === 'Enter' && showDetail(s)}>
          <div class="name">{s.display_name || s.name}</div>
          {#if s.summary}
            <div class="summary">{s.summary}</div>
          {:else if s.description}
            <div class="summary">{s.description}</div>
          {/if}
          <div class="meta">
            <span class="badge-small">{s.tool}</span>
            <span class="badge-small">{s.category}</span>
            <span>{s.size} {_('skills.bytes')}</span>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .toolbar { display:flex; gap:8px; margin-bottom:16px; }
  .toolbar input, .toolbar select { background:#1a1a28; border:1px solid #2a2a3a; color:#ccc; padding:8px 12px; border-radius:6px; font-size:13px; }
  .toolbar input:focus, .toolbar select:focus { border-color:#7c5cfc; outline:none; }
  .skill-grid { display:grid; grid-template-columns:repeat(auto-fill, minmax(280px, 1fr)); gap:12px; }
  .skill-card { background:#1a1a28; border:1px solid #2a2a3a; border-radius:8px; padding:16px; cursor:pointer; transition:.15s; }
  .skill-card:hover { border-color:#7c5cfc; background:#1e1e2e; }
  .skill-card .name { font-size:15px; font-weight:600; margin-bottom:6px; color:#e0e0e0; }
  .skill-card .summary { font-size:12px; color:#888; margin-bottom:10px; line-height:1.4; overflow:hidden; display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; }
  .skill-card .meta { display:flex; gap:8px; font-size:11px; color:#555; align-items:center; }
  .badge-small { background:#2a2a3a; color:#888; padding:1px 6px; border-radius:3px; font-size:10px; }
  .back { background:none; border:none; color:#888; cursor:pointer; font-size:13px; margin-bottom:12px; }
  .back:hover { color:#ccc; }
  .detail-card { background:#1a1a28; border:1px solid #7c5cfc; border-radius:8px; padding:24px; }
  .detail-card h3 { font-size:20px; color:#7c5cfc; margin-bottom:12px; }
  .meta-bar { display:flex; gap:8px; margin-bottom:16px; }
  .badge { background:#2a2a3a; color:#aaa; padding:3px 10px; border-radius:4px; font-size:12px; }
  .desc { color:#aaa; line-height:1.7; margin-bottom:16px; }
  .tags-bar { display:flex; gap:6px; flex-wrap:wrap; }
  .tag { background:#7c5cfc20; color:#7c5cfc; padding:2px 10px; border-radius:4px; font-size:12px; }
</style>