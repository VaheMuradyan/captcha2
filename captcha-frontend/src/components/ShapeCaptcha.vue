<template>
    <div class="flex flex-col items-center gap-4 p-4">
      <h2 class="text-xl font-bold mb-4">Shape CAPTCHA</h2>
      
      <!-- CAPTCHA Grid -->
      <div class="relative" ref="captchaContainer">
        <img 
          :src="captchaImageUrl" 
          alt="CAPTCHA" 
          class="border border-gray-300 rounded"
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
            class="border border-transparent hover:border-blue-400 cursor-pointer relative"
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
          class="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded"
        >
          Refresh
        </button>
        <button 
          @click="verify"
          :disabled="selectedCells.length !== 3"
          class="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded disabled:opacity-50"
        >
          Verify
        </button>
      </div>
  
      <!-- Status Message -->
      <div 
        v-if="message" 
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
        imageWidth: 300,  // Modified for 4x6 aspect ratio
        imageHeight: 200,
      }
    },
    methods: {
      async refreshCaptcha() {
        this.selectedCells = []
        this.message = ''
        this.valid = false
        this.captchaImageUrl = `http://localhost:8080/api/captcha?t=${Date.now()}`
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
          // Remove this cell and any subsequent selections
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
            },
            body: JSON.stringify(this.selectedCells),
          })
          
          const result = await response.json()
          this.message = result.message
          this.valid = result.valid
          
          if (result.valid) {
            this.$emit('captcha-success')
          } else {
            setTimeout(() => this.refreshCaptcha(), 1500)
          }
        } catch (error) {
          this.message = 'Error verifying CAPTCHA'
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