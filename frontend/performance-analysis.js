/**
 * Dashboard Performance Analysis Script
 * Measures React performance optimizations for Issue #37
 * Target: 60fps scroll performance with 100+ repositories
 */

class PerformanceAnalyzer {
  constructor() {
    this.results = {
      timestamp: new Date().toISOString(),
      testConfig: {},
      renderMetrics: {},
      scrollMetrics: {},
      memoryMetrics: {},
      recommendations: []
    };
  }

  /**
   * Initialize performance measurement tools
   */
  async init() {
    // Check if we're in a browser environment
    if (typeof window === 'undefined') {
      throw new Error('Performance analysis must run in browser environment');
    }

    // Ensure React DevTools is available for profiling
    if (window.__REACT_DEVTOOLS_GLOBAL_HOOK__) {
      console.log('✅ React DevTools detected - profiling enabled');
    } else {
      console.warn('⚠️ React DevTools not detected - install for detailed profiling');
    }

    // Check for Performance Observer API
    if ('PerformanceObserver' in window) {
      this.setupPerformanceObserver();
      console.log('✅ Performance Observer API available');
    } else {
      console.warn('⚠️ Performance Observer API not available');
    }

    return this;
  }

  /**
   * Setup Performance Observer for measuring render performance
   */
  setupPerformanceObserver() {
    // Measure long tasks (>50ms) that could cause jank
    if (PerformanceObserver.supportedEntryTypes.includes('longtask')) {
      const longTaskObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          console.warn(`🐌 Long task detected: ${entry.duration}ms`);
          this.results.renderMetrics.longTasks = this.results.renderMetrics.longTasks || [];
          this.results.renderMetrics.longTasks.push({
            duration: entry.duration,
            startTime: entry.startTime
          });
        }
      });
      longTaskObserver.observe({ entryTypes: ['longtask'] });
    }

    // Measure layout thrashing
    if (PerformanceObserver.supportedEntryTypes.includes('layout-shift')) {
      const layoutShiftObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.value > 0.1) { // Significant layout shift
            console.warn(`📐 Layout shift detected: ${entry.value}`);
            this.results.renderMetrics.layoutShifts = this.results.renderMetrics.layoutShifts || [];
            this.results.renderMetrics.layoutShifts.push({
              value: entry.value,
              startTime: entry.startTime
            });
          }
        }
      });
      layoutShiftObserver.observe({ entryTypes: ['layout-shift'] });
    }
  }

  /**
   * Test scroll performance with different repository counts
   */
  async testScrollPerformance(repositoryCount = 100) {
    console.log(`🔬 Testing scroll performance with ${repositoryCount} repositories`);
    
    const startTime = performance.now();
    const frameStats = {
      frames: 0,
      droppedFrames: 0,
      frameRates: []
    };

    // Simulate scroll events and measure frame rate
    return new Promise((resolve) => {
      let lastFrameTime = startTime;
      let animationId;
      
      const measureFrame = (currentTime) => {
        const frameDuration = currentTime - lastFrameTime;
        const fps = 1000 / frameDuration;
        
        frameStats.frames++;
        frameStats.frameRates.push(fps);
        
        if (fps < 50) { // Consider <50fps as dropped frames
          frameStats.droppedFrames++;
        }
        
        lastFrameTime = currentTime;
        
        // Continue for 3 seconds of measurement
        if (currentTime - startTime < 3000) {
          animationId = requestAnimationFrame(measureFrame);
        } else {
          cancelAnimationFrame(animationId);
          
          const avgFps = frameStats.frameRates.reduce((a, b) => a + b, 0) / frameStats.frameRates.length;
          const minFps = Math.min(...frameStats.frameRates);
          const dropRate = (frameStats.droppedFrames / frameStats.frames) * 100;
          
          this.results.scrollMetrics = {
            repositoryCount,
            averageFps: Math.round(avgFps * 10) / 10,
            minimumFps: Math.round(minFps * 10) / 10,
            droppedFramePercentage: Math.round(dropRate * 10) / 10,
            totalFrames: frameStats.frames,
            testDuration: currentTime - startTime
          };
          
          console.log(`📊 Scroll Performance Results:
            - Average FPS: ${this.results.scrollMetrics.averageFps}
            - Minimum FPS: ${this.results.scrollMetrics.minimumFps}
            - Dropped Frame Rate: ${this.results.scrollMetrics.droppedFramePercentage}%
            - Target: 60 FPS ${avgFps >= 60 ? '✅' : '❌'}
          `);
          
          resolve(this.results.scrollMetrics);
        }
      };
      
      // Start measurement
      animationId = requestAnimationFrame(measureFrame);
      
      // Simulate scroll events
      this.simulateScrolling(3000);
    });
  }

  /**
   * Simulate scrolling behavior for testing
   */
  simulateScrolling(duration) {
    const container = document.querySelector('[data-testid="repository-list"]') || 
                    document.querySelector('.grid') || 
                    document.documentElement;
    
    if (!container) {
      console.warn('⚠️ Could not find scrollable container for testing');
      return;
    }

    const startTime = Date.now();
    const scrollHeight = container.scrollHeight || document.body.scrollHeight;
    
    const scrollInterval = setInterval(() => {
      const elapsed = Date.now() - startTime;
      const progress = elapsed / duration;
      
      if (progress >= 1) {
        clearInterval(scrollInterval);
        return;
      }
      
      // Smooth scrolling simulation
      const scrollPosition = Math.sin(progress * Math.PI * 4) * (scrollHeight / 4) + (scrollHeight / 2);
      container.scrollTop = scrollPosition;
    }, 16); // ~60fps
  }

  /**
   * Measure memory usage and potential leaks
   */
  async measureMemoryUsage() {
    console.log('🧠 Measuring memory usage');
    
    if ('memory' in performance) {
      const memory = performance.memory;
      this.results.memoryMetrics = {
        usedJSHeapSize: Math.round(memory.usedJSHeapSize / 1048576), // MB
        totalJSHeapSize: Math.round(memory.totalJSHeapSize / 1048576), // MB
        jsHeapSizeLimit: Math.round(memory.jsHeapSizeLimit / 1048576), // MB
        heapUsagePercentage: Math.round((memory.usedJSHeapSize / memory.jsHeapSizeLimit) * 100)
      };
      
      console.log(`📊 Memory Usage:
        - Used Heap: ${this.results.memoryMetrics.usedJSHeapSize}MB
        - Total Heap: ${this.results.memoryMetrics.totalJSHeapSize}MB
        - Heap Limit: ${this.results.memoryMetrics.jsHeapSizeLimit}MB
        - Usage: ${this.results.memoryMetrics.heapUsagePercentage}%
      `);
    } else {
      console.warn('⚠️ Memory API not available in this browser');
      this.results.memoryMetrics = { error: 'Memory API not available' };
    }
    
    return this.results.memoryMetrics;
  }

  /**
   * Analyze React component render performance
   */
  async analyzeRenderPerformance() {
    console.log('⚛️ Analyzing React render performance');
    
    // Check for React-specific performance markers
    const reactMarkers = performance.getEntriesByType('measure')
      .filter(entry => entry.name.includes('⚛️') || entry.name.includes('React'));
    
    if (reactMarkers.length > 0) {
      this.results.renderMetrics.reactMeasures = reactMarkers.map(marker => ({
        name: marker.name,
        duration: marker.duration,
        startTime: marker.startTime
      }));
      
      console.log(`📊 Found ${reactMarkers.length} React performance markers`);
    }

    // Analyze DOM mutations which could indicate unnecessary re-renders
    this.analyzeDOMMutations();
    
    return this.results.renderMetrics;
  }

  /**
   * Monitor DOM mutations to detect unnecessary re-renders
   */
  analyzeDOMMutations() {
    let mutationCount = 0;
    const startTime = Date.now();
    
    const observer = new MutationObserver((mutations) => {
      mutationCount += mutations.length;
    });
    
    observer.observe(document.body, {
      childList: true,
      subtree: true,
      attributes: true
    });
    
    // Stop observing after 5 seconds
    setTimeout(() => {
      observer.disconnect();
      const duration = Date.now() - startTime;
      const mutationsPerSecond = mutationCount / (duration / 1000);
      
      this.results.renderMetrics.domMutations = {
        count: mutationCount,
        duration: duration,
        mutationsPerSecond: Math.round(mutationsPerSecond * 10) / 10
      };
      
      console.log(`📊 DOM Mutations: ${mutationCount} in ${duration}ms (${mutationsPerSecond}/sec)`);
      
      if (mutationsPerSecond > 100) {
        console.warn('⚠️ High DOM mutation rate detected - potential performance issue');
      }
    }, 5000);
  }

  /**
   * Generate performance recommendations
   */
  generateRecommendations() {
    const recommendations = [];
    
    // Scroll performance recommendations
    if (this.results.scrollMetrics) {
      const { averageFps, droppedFramePercentage } = this.results.scrollMetrics;
      
      if (averageFps < 60) {
        recommendations.push({
          category: 'Scroll Performance',
          severity: 'high',
          issue: `Average FPS (${averageFps}) below 60fps target`,
          solution: 'Implement virtual scrolling for large repository lists'
        });
      }
      
      if (droppedFramePercentage > 10) {
        recommendations.push({
          category: 'Scroll Performance',
          severity: 'medium',
          issue: `${droppedFramePercentage}% dropped frames during scrolling`,
          solution: 'Optimize RepositoryCard rendering with additional memoization'
        });
      }
    }
    
    // Memory recommendations
    if (this.results.memoryMetrics && this.results.memoryMetrics.heapUsagePercentage > 70) {
      recommendations.push({
        category: 'Memory Usage',
        severity: 'medium',
        issue: `High memory usage: ${this.results.memoryMetrics.heapUsagePercentage}%`,
        solution: 'Implement component cleanup and memory leak prevention'
      });
    }
    
    // Long task recommendations
    if (this.results.renderMetrics.longTasks && this.results.renderMetrics.longTasks.length > 0) {
      recommendations.push({
        category: 'Render Performance',
        severity: 'high',
        issue: `${this.results.renderMetrics.longTasks.length} long tasks detected`,
        solution: 'Break down large operations using time slicing or web workers'
      });
    }
    
    // Layout shift recommendations
    if (this.results.renderMetrics.layoutShifts && this.results.renderMetrics.layoutShifts.length > 0) {
      recommendations.push({
        category: 'User Experience',
        severity: 'medium',
        issue: 'Layout shifts detected during rendering',
        solution: 'Implement skeleton loaders and reserve space for dynamic content'
      });
    }
    
    this.results.recommendations = recommendations;
    return recommendations;
  }

  /**
   * Run comprehensive performance analysis
   */
  async runFullAnalysis(repositoryCount = 100) {
    console.log('🚀 Starting comprehensive dashboard performance analysis');
    
    try {
      await this.init();
      
      this.results.testConfig = {
        repositoryCount,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
        viewport: {
          width: window.innerWidth,
          height: window.innerHeight
        }
      };
      
      // Run all performance tests
      await this.measureMemoryUsage();
      await this.analyzeRenderPerformance();
      await this.testScrollPerformance(repositoryCount);
      
      // Generate recommendations
      this.generateRecommendations();
      
      console.log('✅ Performance analysis complete');
      console.log('📊 Full Results:', this.results);
      
      return this.results;
      
    } catch (error) {
      console.error('❌ Performance analysis failed:', error);
      throw error;
    }
  }

  /**
   * Export results to JSON for further analysis
   */
  exportResults() {
    const blob = new Blob([JSON.stringify(this.results, null, 2)], {
      type: 'application/json'
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `dashboard-performance-${Date.now()}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }
}

// Export for use in browser console or testing
window.PerformanceAnalyzer = PerformanceAnalyzer;

// Auto-run analysis if script is loaded directly
if (typeof module === 'undefined') {
  console.log('🔬 Dashboard Performance Analyzer loaded');
  console.log('Run: new PerformanceAnalyzer().runFullAnalysis(100)');
}