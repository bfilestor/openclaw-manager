<template>
  <div>
    <el-space class="diff-toolbar" wrap>
      <el-text type="info">{{ t('diffViewer.mode') }}</el-text>
      <el-radio-group v-model="viewMode" size="small">
        <el-radio-button label="unified">{{ t('diffViewer.unified') }}</el-radio-button>
        <el-radio-button label="split">{{ t('diffViewer.split') }}</el-radio-button>
      </el-radio-group>
    </el-space>

    <el-scrollbar :height="height">
      <pre v-if="viewMode === 'unified'" class="revision-content diff-content"><template v-for="(line, idx) in diffLines" :key="idx"><span :class="line.type">{{ line.text }}
</span></template></pre>

      <table v-else class="split-diff-table">
        <thead>
          <tr>
            <th>{{ resolvedLeftTitle }}</th>
            <th>{{ resolvedRightTitle }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, idx) in splitDiffRows" :key="idx">
            <td :class="['left', row.leftType]">{{ row.left }}</td>
            <td :class="['right', row.rightType]">{{ row.right }}</td>
          </tr>
        </tbody>
      </table>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { diffLines as calcDiffLines } from 'diff'
import { useI18n } from 'vue-i18n'

const props = withDefaults(defineProps<{
  fromText: string
  toText: string
  leftTitle?: string
  rightTitle?: string
  height?: number | string
}>(), {
  leftTitle: '',
  rightTitle: '',
  height: 460,
})

const { t } = useI18n()
const viewMode = ref<'unified' | 'split'>('unified')
const resolvedLeftTitle = computed(() => props.leftTitle || t('diffViewer.leftTitle'))
const resolvedRightTitle = computed(() => props.rightTitle || t('diffViewer.rightTitle'))

watch(() => [props.fromText, props.toText], () => {
  viewMode.value = 'unified'
})

const diffLines = computed(() => {
  const parts = calcDiffLines(props.fromText || '', props.toText || '')
  return parts.flatMap((part) => {
    const rows = String(part.value || '').split('\n')
    if (rows.length > 0 && rows[rows.length - 1] === '') rows.pop()
    const type: 'same' | 'add' | 'remove' = part.added ? 'add' : part.removed ? 'remove' : 'same'
    return rows.map((row) => ({
      type,
      text: `${type === 'add' ? '+' : type === 'remove' ? '-' : ' '} ${row}`,
    }))
  })
})

const splitDiffRows = computed(() => {
  const parts = calcDiffLines(props.fromText || '', props.toText || '')
  const rows: { left: string; right: string; leftType: 'same' | 'add' | 'remove'; rightType: 'same' | 'add' | 'remove' }[] = []

  for (let i = 0; i < parts.length; i++) {
    const part = parts[i]
    if (part.removed && parts[i + 1]?.added) {
      const removedLines = String(part.value || '').split('\n')
      const addedLines = String(parts[i + 1].value || '').split('\n')
      if (removedLines.length > 0 && removedLines[removedLines.length - 1] === '') removedLines.pop()
      if (addedLines.length > 0 && addedLines[addedLines.length - 1] === '') addedLines.pop()
      const maxLen = Math.max(removedLines.length, addedLines.length)
      for (let j = 0; j < maxLen; j++) {
        rows.push({
          left: removedLines[j] ?? '',
          right: addedLines[j] ?? '',
          leftType: 'remove',
          rightType: 'add',
        })
      }
      i += 1
      continue
    }

    const lines = String(part.value || '').split('\n')
    if (lines.length > 0 && lines[lines.length - 1] === '') lines.pop()
    if (part.added) {
      lines.forEach((line) => rows.push({ left: '', right: line, leftType: 'same', rightType: 'add' }))
    } else if (part.removed) {
      lines.forEach((line) => rows.push({ left: line, right: '', leftType: 'remove', rightType: 'same' }))
    } else {
      lines.forEach((line) => rows.push({ left: line, right: line, leftType: 'same', rightType: 'same' }))
    }
  }

  return rows
})
</script>

<style scoped>
.diff-toolbar {
  margin-bottom: 8px;
}
.revision-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
}
.diff-content .add {
  background: #ecfdf3;
  color: #1b5e20;
}
.diff-content .remove {
  background: #fff1f0;
  color: #b42318;
}
.diff-content .same {
  color: #667085;
}
.split-diff-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
}
.split-diff-table th,
.split-diff-table td {
  border: 1px solid #eaecf0;
  padding: 4px 8px;
  vertical-align: top;
  white-space: pre-wrap;
  word-break: break-word;
}
.split-diff-table td.add {
  background: #ecfdf3;
  color: #1b5e20;
}
.split-diff-table td.remove {
  background: #fff1f0;
  color: #b42318;
}
.split-diff-table td.same {
  color: #667085;
}
</style>
