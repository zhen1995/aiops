<template>
  <div class="markdown-body" v-html="sanitizedHtml"></div>
</template>

<script setup>
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps({
  content: {
    type: String,
    default: ''
  }
})

const sanitizedHtml = computed(() => {
  if (!props.content) return ''
  const raw = marked.parse(props.content, { async: false })
  return DOMPurify.sanitize(raw)
})
</script>

<style scoped>
.markdown-body {
  font-size: 13.5px;
  line-height: 1.7;
  color: var(--c-text);
}

.markdown-body :deep(p) {
  margin: 0.6em 0;
}

.markdown-body :deep(p:first-child) {
  margin-top: 0;
}

.markdown-body :deep(p:last-child) {
  margin-bottom: 0;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin: 1em 0 0.5em;
  font-weight: 600;
  line-height: 1.4;
}

.markdown-body :deep(h1) { font-size: 1.5em; }
.markdown-body :deep(h2) { font-size: 1.3em; }
.markdown-body :deep(h3) { font-size: 1.15em; }

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 1.5em;
  margin: 0.5em 0;
}

.markdown-body :deep(li) {
  margin: 0.25em 0;
}

.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 0.8em 0;
  font-size: 13px;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid var(--c-border);
  padding: 8px 10px;
  text-align: left;
}

.markdown-body :deep(th) {
  background: var(--c-bg);
  font-weight: 600;
}

.markdown-body :deep(tr:nth-child(even)) {
  background: var(--c-bg);
}

.markdown-body :deep(code) {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
  background: var(--c-bg);
  padding: 2px 5px;
  border-radius: 4px;
  font-size: 0.92em;
  color: var(--c-danger, #c93b3b);
}

.markdown-body :deep(pre) {
  background: #f6f8fa;
  border: 1px solid var(--c-border);
  border-radius: 8px;
  padding: 12px 14px;
  overflow-x: auto;
  margin: 0.8em 0;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
  border-radius: 0;
  color: inherit;
  font-size: 13px;
}

.markdown-body :deep(blockquote) {
  margin: 0.8em 0;
  padding: 0 1em;
  color: var(--c-text-2);
  border-left: 4px solid var(--c-border);
}

.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid var(--c-border);
  margin: 1em 0;
}

.markdown-body :deep(a) {
  color: var(--c-primary);
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}
</style>
