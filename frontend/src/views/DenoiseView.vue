<template>
  <div>
    <PageHeader title="告警降噪" desc="多级降噪策略链 · 告警压缩 86.4% · 人工反馈闭环持续优化">
      <button class="btn">近 24 小时</button>
      <button class="btn btn-primary">导出降噪报告</button>
    </PageHeader>

    <!-- 统计卡片 -->
    <div class="kpi-grid">
      <StatCard v-for="k in kpiCards" :key="k.label" v-bind="k" />
    </div>

    <!-- 降噪漏斗 -->
    <div class="card funnel-card">
      <h3 class="card-title">七级降噪漏斗</h3>
      <p class="card-sub">今日 2,316 条原始告警经逐级过滤，最终仅 156 条有效通知触达值班人员</p>
      <ChartBox :option="funnelOption" height="340px" />
    </div>

    <!-- 降噪策略 -->
    <div class="card">
      <h3 class="card-title">降噪策略</h3>
      <p class="card-sub">五项策略按序执行，可随时启停；停用策略的拦截量不再计入压缩率</p>
      <div class="policy-grid">
        <div v-for="p in policies" :key="p.name" class="policy-card" :class="{ off: !p.enabled }">
          <div class="policy-head">
            <span class="policy-name">{{ p.name }}</span>
            <label class="switch" :title="p.enabled ? '点击停用' : '点击启用'">
              <input type="checkbox" v-model="p.enabled" />
              <span class="slider"></span>
            </label>
          </div>
          <p class="policy-desc">{{ p.desc }}</p>
          <p class="policy-effect">{{ p.effect }}</p>
          <div class="policy-foot">
            <span class="muted">今日已拦截</span>
            <span class="policy-reduced mono">{{ p.reduced.toLocaleString() }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 生命周期与反馈闭环 -->
    <div class="card lifecycle-card">
      <h3 class="card-title">告警生命周期与反馈闭环</h3>
      <p class="card-sub">降噪并非一次性过滤，而是随人工反馈持续进化的闭环系统</p>
      <div class="lifecycle">
        <div v-for="(s, i) in lifecycle" :key="s.title" class="life-step">
          <span class="life-no">{{ i + 1 }}</span>
          <div>
            <p class="life-title">{{ s.title }}</p>
            <p class="life-desc">{{ s.desc }}</p>
          </div>
        </div>
      </div>
      <p class="life-note">
        值班人员对每条通知标记「有效 / 误报 / 重复」，反馈样本回流至相似度模型与动态阈值模块，
        每日凌晨自动重训练。近 30 天误报率由 12.6% 降至 4.1%，告警压缩率稳定在 80% 目标线以上。
      </p>
    </div>
  </div>
</template>

<script setup>
import { computed, ref } from 'vue'
import PageHeader from '../components/PageHeader.vue'
import StatCard from '../components/StatCard.vue'
import ChartBox from '../components/ChartBox.vue'
import { denoisePolicies, denoiseStats } from '../mock/data'

// 本地响应式副本，驱动开关交互
const policies = ref(denoisePolicies.map((p) => ({ ...p })))

const blockedTotal = computed(() =>
  policies.value.filter((p) => p.enabled).reduce((sum, p) => sum + p.reduced, 0)
)

const kpiCards = computed(() => [
  { label: '原始告警总量', value: denoiseStats.rawTotal.toLocaleString(), delta: '+12%', deltaType: 'down', hint: '较昨日 2,068 条' },
  { label: '有效告警', value: denoiseStats.final, delta: '-86.4%', deltaType: 'up', hint: '实际触达值班' },
  { label: '压缩率', value: '86.4', unit: '%', delta: '+4.2%', deltaType: 'up', hint: '目标 ≥80%' },
  { label: '已拦截告警', value: blockedTotal.value.toLocaleString(), delta: '5 项策略', deltaType: 'flat', hint: '启用中策略合计' }
])

// 浅青 → 深青 层次配色
const funnelColors = ['#e7f3f1', '#c4e2de', '#9bd0ca', '#6fb9b1', '#43a298', '#22897f', '#0a5f57']

const funnelOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}：{c} 条' },
  series: [{
    type: 'funnel',
    left: '12%', right: '12%', top: 8, bottom: 8,
    minSize: '18%', maxSize: '100%',
    sort: 'descending',
    gap: 4,
    label: { show: true, position: 'inside', formatter: '{b}  {c}', fontSize: 12, color: '#20302d' },
    itemStyle: { borderColor: '#fff', borderWidth: 1 },
    emphasis: { label: { fontSize: 13, fontWeight: 600 } },
    data: denoiseStats.funnel.map((f, i) => ({
      ...f,
      itemStyle: { color: funnelColors[i] },
      label: { color: i >= 4 ? '#fff' : '#20302d' }
    }))
  }]
}))

const lifecycle = [
  { title: '告警产生', desc: '检测引擎与静态规则实时触发原始告警，统一进入降噪管道' },
  { title: '多级降噪', desc: '规则过滤 → 窗口聚合 → 相似度去重 → 拓扑抑制 → 动态阈值 逐级压缩' },
  { title: '智能分级', desc: '按影响面与紧急度自动定级 P0-P4，决定通知方式与升级策略' },
  { title: '通知触达', desc: '按值班表路由至电话 / 短信 / IM，超时未确认自动升级' },
  { title: '人工反馈', desc: '值班人员标记有效性，误报与漏报样本沉淀为训练数据' },
  { title: '模型优化', desc: '反馈样本每日重训练相似度与阈值模型，降噪效果持续提升' }
]
</script>

<style scoped>
.kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 16px; }
.funnel-card { margin-bottom: 16px; }
.policy-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(210px, 1fr)); gap: 14px; }
.policy-card {
  border: 1px solid var(--c-border);
  border-radius: var(--radius-card);
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: opacity 0.2s ease, filter 0.2s ease;
}
.policy-card.off { opacity: 0.55; filter: grayscale(0.6); }
.policy-head { display: flex; justify-content: space-between; align-items: center; }
.policy-name { font-size: 14px; font-weight: 600; }
.policy-desc { margin: 0; font-size: 12.5px; color: var(--c-text-2); line-height: 1.5; }
.policy-effect { margin: 0; font-size: 12.5px; font-weight: 600; color: var(--c-primary); }
.policy-foot { margin-top: auto; display: flex; justify-content: space-between; align-items: center; border-top: 1px dashed var(--c-border); padding-top: 8px; }
.policy-reduced { font-size: 16px; font-weight: 700; color: var(--c-primary-dark); }

/* 开关 */
.switch { position: relative; width: 38px; height: 20px; flex-shrink: 0; cursor: pointer; }
.switch input { opacity: 0; width: 0; height: 0; }
.slider { position: absolute; inset: 0; border-radius: 20px; background: var(--c-p4); transition: background 0.2s; }
.slider::before {
  content: ''; position: absolute; left: 3px; top: 3px;
  width: 14px; height: 14px; border-radius: 50%;
  background: #fff; transition: transform 0.2s;
}
.switch input:checked + .slider { background: var(--c-primary); }
.switch input:checked + .slider::before { transform: translateX(18px); }

.lifecycle-card { margin-top: 16px; }
.lifecycle { display: grid; grid-template-columns: repeat(3, 1fr); gap: 14px 20px; margin-bottom: 14px; }
.life-step { display: flex; gap: 10px; align-items: flex-start; padding: 10px 12px; border-radius: var(--radius-card); background: var(--c-primary-soft); border: 1px solid var(--c-border); }
.life-no {
  flex-shrink: 0; width: 22px; height: 22px; border-radius: 50%;
  background: var(--c-primary); color: #fff; font-size: 12px; font-weight: 700;
  display: flex; align-items: center; justify-content: center; margin-top: 2px;
}
.life-title { margin: 0 0 2px; font-size: 13px; font-weight: 600; }
.life-desc { margin: 0; font-size: 12px; color: var(--c-text-2); line-height: 1.5; }
.life-note { margin: 0; font-size: 12.5px; color: var(--c-text-2); background: var(--c-primary-tint); border-radius: var(--radius-card); padding: 10px 14px; line-height: 1.7; }

@media (max-width: 1200px) {
  .kpi-grid { grid-template-columns: repeat(2, 1fr); }
  .lifecycle { grid-template-columns: 1fr; }
}
</style>
