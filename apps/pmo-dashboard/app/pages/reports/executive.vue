<script setup lang="ts">
import type { ExecutiveReport } from '~/types/api'

definePageMeta({ middleware: 'auth', requiredFunction: 'view_executive_report' })
useHead({ title: 'エグゼクティブレポート — pmotto' })

const api = useApi()
const reports = ref<ExecutiveReport[]>([])
const errorMsg = ref('')
const loading = ref(true)

// 一覧は新しい順。既定で最新の1件を開いた状態にする。
const selectedId = ref<number | null>(null)
const selected = computed(() => reports.value.find((r) => r.id === selectedId.value) ?? null)

async function load() {
  try {
    const res = await api<{ reports: ExecutiveReport[] }>('/reports/executive')
    reports.value = res.reports ?? []
    selectedId.value = reports.value[0]?.id ?? null
  } catch (e) {
    errorMsg.value = apiError(e, 'エグゼクティブレポートの取得に失敗しました')
  } finally {
    loading.value = false
  }
}
onMounted(load)

function percent(n: number): string {
  return `${Math.round(n * 10) / 10}%`
}

// progress_trend は前週差。増減が一目で分かるよう符号を明示する。
function trend(n: number): string {
  if (n === 0) return '±0'
  return `${n > 0 ? '+' : ''}${Math.round(n * 10) / 10}`
}

function trendClass(n: number): string {
  if (n > 0) return 'text-ink'
  if (n < 0) return 'text-grad-coral'
  return 'text-ink-muted'
}

function reportLabel(r: ExecutiveReport): string {
  return `${r.report_date.slice(0, 10)}（${r.report_type}）`
}
</script>

<template>
  <AppShell>
    <div>
      <p class="eyebrow text-ink-muted">Executive</p>
      <h1 class="display-md mt-3 text-ink">エグゼクティブレポート</h1>
      <p class="mt-3 text-sm text-ink-muted">
        n8n の週次ワークフローが全プロジェクトを横断集計し、AI が経営視点のコメントを付与したものです。
      </p>
    </div>

    <p v-if="errorMsg" class="mt-4 text-sm text-grad-coral" role="alert">{{ errorMsg }}</p>

    <p v-else-if="loading" class="mt-8 text-sm text-ink-muted">読み込み中…</p>

    <!-- レポートは n8n が生成するため、ワークフロー未実行だと0件になる。
         画面の不具合と切り分けられるよう、その旨を明示する。 -->
    <div v-else-if="reports.length === 0" class="mt-8 rounded-xl bg-surface-1 p-6 ring-1 ring-hairline">
      <p class="text-sm text-ink">レポートがまだありません。</p>
      <p class="mt-2 text-sm text-ink-muted">
        n8n の週次ワークフロー（weekly_workflow）を実行すると、ここに表示されます。
      </p>
    </div>

    <template v-else>
      <label class="mt-8 flex max-w-sm flex-col gap-2">
        <span class="text-sm text-ink-muted">レポート週</span>
        <select
          v-model="selectedId"
          class="rounded-lg bg-surface-1 px-4 py-2 text-ink ring-1 ring-hairline focus:outline-none focus:ring-2 focus:ring-accent-blue"
        >
          <option v-for="r in reports" :key="r.id" :value="r.id">{{ reportLabel(r) }}</option>
        </select>
      </label>

      <template v-if="selected">
        <div class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="rounded-xl bg-surface-1 p-5 ring-1 ring-hairline">
            <p class="text-sm text-ink-muted">管理プロジェクト</p>
            <p class="mt-2 text-2xl text-ink">{{ selected.content.total_projects }}</p>
          </div>
          <div class="rounded-xl bg-surface-1 p-5 ring-1 ring-hairline">
            <p class="text-sm text-ink-muted">順調</p>
            <p class="mt-2 text-2xl text-ink">{{ selected.content.on_track_count }}</p>
          </div>
          <div class="rounded-xl bg-surface-1 p-5 ring-1 ring-hairline">
            <p class="text-sm text-ink-muted">高リスク</p>
            <p class="mt-2 text-2xl text-ink">{{ selected.content.high_risk_count }}</p>
          </div>
          <div class="rounded-xl bg-surface-1 p-5 ring-1 ring-hairline">
            <p class="text-sm text-ink-muted">平均進捗率</p>
            <p class="mt-2 text-2xl text-ink">{{ percent(selected.content.average_progress) }}</p>
          </div>
        </div>

        <section v-if="selected.ai_comment" class="mt-8 rounded-xl bg-surface-2 p-6 ring-1 ring-hairline">
          <p class="eyebrow text-ink-muted">AI コメント</p>
          <p class="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-ink">{{ selected.ai_comment }}</p>
        </section>
        <!-- ai_comment は AI 生成ステップが後から UPDATE で埋めるため、
             集計直後は空になりうる。欠落ではないことを示す。 -->
        <p v-else class="mt-8 text-sm text-ink-muted">AI コメントはまだ生成されていません。</p>

        <section v-if="selected.content.executive_summary" class="mt-6">
          <p class="eyebrow text-ink-muted">概況</p>
          <p class="mt-3 text-sm leading-relaxed text-ink">{{ selected.content.executive_summary }}</p>
        </section>

        <section class="mt-10">
          <h2 class="text-lg text-ink">プロジェクト別</h2>
          <div class="mt-4 overflow-x-auto rounded-xl ring-1 ring-hairline">
            <table class="w-full text-left text-sm">
              <thead class="bg-surface-1 text-ink-muted">
                <tr>
                  <th class="px-4 py-3 font-medium">プロジェクト</th>
                  <th class="px-4 py-3 font-medium">進捗率</th>
                  <th class="px-4 py-3 font-medium">前週差</th>
                  <th class="px-4 py-3 font-medium">高リスク</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="p in selected.content.projects"
                  :key="p.project_id"
                  class="border-t border-hairline"
                >
                  <td class="px-4 py-3 text-ink">
                    <NuxtLink :to="`/projects/${p.project_id}`" class="hover:text-accent-blue">
                      {{ p.project_name }}
                    </NuxtLink>
                  </td>
                  <td class="px-4 py-3 text-ink">{{ percent(p.avg_progress) }}</td>
                  <td class="px-4 py-3" :class="trendClass(p.progress_trend)">{{ trend(p.progress_trend) }}</td>
                  <td class="px-4 py-3 text-ink">{{ p.high_risk_count }}</td>
                </tr>
                <tr v-if="selected.content.projects.length === 0">
                  <td class="px-4 py-6 text-sm text-ink-muted" colspan="4">対象プロジェクトがありません。</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </template>
    </template>
  </AppShell>
</template>
