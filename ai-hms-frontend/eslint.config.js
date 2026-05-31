import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: "Literal[value=/\\brounded-(xl|2xl|3xl)\\b/]",
          message: '禁止 rounded-xl/2xl/3xl，请改用 rounded-md / rounded-lg（见 docs/ui-ux-plan/U0-design-system.md）',
        },
        {
          selector: "TemplateElement[value.raw=/\\brounded-(xl|2xl|3xl)\\b/]",
          message: '禁止 rounded-xl/2xl/3xl（模板字符串），请改用 rounded-md / rounded-lg',
        },
        {
          selector: "Literal[value=/text-\\[(10|11)px\\]/]",
          message: '禁止裸 text-[10px]/[11px]，请改用 text-meta，或在白名单场景使用 text-density-strict',
        },
        {
          selector: "TemplateElement[value.raw=/text-\\[(10|11)px\\]/]",
          message: '禁止裸 text-[10px]/[11px]（模板字符串）',
        },
        {
          selector: "Literal[value=/\\b!important\\b/]",
          message: '禁止在业务页面新增 !important，主题适配请用 surface/foreground token',
        },
      ],
    },
  },
  // U5-Step4: 逐文件清理后移出豁免清单
  {
    files: [
      'src/pages/DictConfig.tsx',
      'src/pages/DeviceManagement.tsx',
      'src/pages/Inventory.tsx',
      'src/pages/MasterData.tsx',
      'src/pages/Settings.tsx',
      'src/pages/Statistics.tsx',
      'src/pages/TreatmentConfig.tsx',
      'src/pages/TreatmentConfig/**',
      'src/pages/RoleManagement.tsx',
      'src/pages/RoleSelect.tsx',
      'src/pages/EducationManagement.tsx',
      'src/pages/BedManagement.tsx',
      'src/pages/WardManagement.tsx',
      'src/pages/WardOverview.tsx',
      'src/pages/UserManagement.tsx',
      'src/pages/Login.tsx',
      'src/pages/dialysis-processing/**',
      'src/pages/dialysis-processing/execution/**',
      'src/pages/patient-detail/**',
      'src/pages/PatientDetail.tsx',
      'src/pages/ScheduleTemplateEditor.tsx',
      'src/pages/ScheduleTemplateList.tsx',
      'src/components/**',
    ],
    rules: { 'no-restricted-syntax': 'off' },
  },
])
