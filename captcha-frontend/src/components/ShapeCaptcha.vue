```vue
<template>
  <div class="flex flex-col items-center gap-4 p-4">
    <h2 class="text-xl font-bold mb-4">Նկարների CAPTCHA</h2>

    <!-- Improved Sequence Display -->
    <div v-if="sequence" class="text-lg mb-4 font-medium text-gray-700 bg-blue-50 p-4 rounded-lg w-full max-w-md">
      <div class="text-center mb-2">Ընտրեք պատկերները հետևյալ հերթականությամբ՝</div>
      <div class="text-center font-bold text-blue-600">
        {{ sequence }}
      </div>
      <div class="text-sm text-gray-600 mt-2 text-center">
        (Սեղմեք պատկերների վրա նշված հերթականությամբ)
      </div>
    </div>
    
    <!-- CAPTCHA Grid -->
    <div class="relative" ref="captchaContainer">
      <img 
        :src="captchaImageUrl" 
        alt="CAPTCHA" 
        class="rounded shadow-md"
        @load="setupGrid"
      />
      
      <!-- Grid Overlay -->
      <div 
        class="absolute top-0 left-0 grid grid-cols-6 grid-rows-4"
        :style="{ 
          width: `${imageWidth}px`, 
          height: `${imageHeight}px`,
          aspectRatio: '3/2'
        }"
      >
        <div
          v-for="i in 24"
          :key="i"
          class="cursor-pointer relative hover:bg-blue-100 hover:bg-opacity-20 transition-colors duration-200"
          @click="handleCellClick(Math.floor((i-1)/6), (i-1)%6)"
        >
          <div 
            v-if="isSelected(Math.floor((i-1)/6), (i-1)%6)"
            class="absolute inset-0 flex items-center justify-center text-xl font-bold text-blue-500 bg-blue-100 bg-opacity-50"
          >
            {{ getSelectionOrder(Math.floor((i-1)/6), (i-1)%6) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Controls -->
    <div class="flex gap-4">
      <button 
        @click="refreshCaptcha"
        class="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded transition-colors duration-200"
      >
        Թարմացնել
      </button>
      <button 
        @click="verify"
        :disabled="selectedCells.length !== 3"
        class="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded disabled:opacity-50 transition-colors duration-200"
      >
        Ստուգել
      </button>
    </div>

    <!-- Status Message -->
    <div 
      v-if="message" 
      class="mt-2 text-center font-medium"
      :class="{
        'text-green-600': valid,
        'text-red-600': !valid && message
      }"
    >
      {{ message }}
    </div>
  </div>
</template>

<script>
export default {
  name: 'ShapeCaptcha',
  data() {
    return {
      captchaImageUrl: '',
      selectedCells: [],
      message: '',
      valid: false,
      imageWidth: 300,
      imageHeight: 200,
      sequence: '',
      captchaKey: '',
    }
  },
  methods: {
    async refreshCaptcha() {
      this.selectedCells = []
      this.message = ''
      this.valid = false
      
      try {
        const response = await fetch('http://localhost:8080/api/captcha')
        if (!response.ok) throw new Error('Failed to fetch CAPTCHA')

        this.captchaKey = response.headers.get('X-Captcha-Key')
        
        const blob = await response.blob()
        this.captchaImageUrl = URL.createObjectURL(blob)

        await this.getSequence()
      } catch (error) {
        console.error('Error fetching CAPTCHA:', error)
        this.message = 'Սխալ CAPTCHA-ի բեռնման ժամանակ'
      }
    },

    async getSequence() {
      if (!this.captchaKey) return
      
      try {
        const response = await fetch('http://localhost:8080/api/sequence', {
          headers: {
            'X-Captcha-Key': this.captchaKey
          }
        })
        
        if (!response.ok) throw new Error('Failed to fetch sequence')
        
        const data = await response.json()
        this.sequence = data.sequence
      } catch (error) {
        console.error('Error fetching sequence:', error)
      }
    },
    
    setupGrid() {
      if (this.$refs.captchaContainer) {
        const img = this.$refs.captchaContainer.querySelector('img')
        this.imageWidth = img.width
        this.imageHeight = img.height
      }
    },
    
    handleCellClick(row, col) {
      const index = this.selectedCells.findIndex(cell => 
        cell[0] === row && cell[1] === col
      )
      
      if (index !== -1) {
        this.selectedCells = this.selectedCells.slice(0, index)
      } else if (this.selectedCells.length < 3) {
        this.selectedCells.push([row, col])
      }
    },
    
    isSelected(row, col) {
      return this.selectedCells.some(cell => 
        cell[0] === row && cell[1] === col
      )
    },
    
    getSelectionOrder(row, col) {
      const index = this.selectedCells.findIndex(cell => 
        cell[0] === row && cell[1] === col
      )
      return index !== -1 ? index + 1 : ''
    },
    
    async verify() {
      if (this.selectedCells.length !== 3) return
      
      try {
        const response = await fetch('http://localhost:8080/api/verify', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-Captcha-Key': this.captchaKey
          },
          body: JSON.stringify(this.selectedCells),
        })
        
        const result = await response.json()
        this.message = result.message
        this.valid = result.valid
        
        if (result.valid) {
          this.$emit('captcha-success')
          this.message = 'Ճիշտ է!'
        } else {
          this.message = 'Սխալ է։ Փորձեք կրկին։'
          setTimeout(() => this.refreshCaptcha(), 1500)
        }
      } catch (error) {
        console.error('Error verifying CAPTCHA:', error)
        this.message = 'Սխալ ստուգման ժամանակ'
        this.valid = false
      }
    }
  },
  mounted() {
    this.refreshCaptcha()
  }
}
</script>

<style scoped>
.grid > div {
  position: relative;
}
.grid > div::before {
  content: '';
  display: block;
  padding-top: 100%;
}
</style>
```