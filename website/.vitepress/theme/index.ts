import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import './custom.css'
import HomeLayout from './HomeLayout.vue'

export default {
  extends: DefaultTheme,
  Layout: HomeLayout,
} satisfies Theme
