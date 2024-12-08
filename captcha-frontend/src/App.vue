<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center">
    <div class="bg-white p-8 rounded-lg shadow-md">
      <ShapeCaptcha @captcha-success="handleCaptchaSuccess" />
    </div>
  </div>
</template>

<script>
import ShapeCaptcha from './components/ShapeCaptcha.vue'

export default {
  name: 'App',
  components: {
    ShapeCaptcha
  },
  methods: {
    handleCaptchaSuccess() {
      alert('CAPTCHA verified successfully!')
    }
  }
}
</script>