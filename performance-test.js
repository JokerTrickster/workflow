/**
 * Automated Dashboard Performance Test Suite
 * Uses browser automation to measure real-world performance
 */

const fs = require('fs').promises;
const path = require('path');

class DashboardPerformanceTest {
  constructor(browserType = 'chromium') {
    this.browserType = browserType;
    this.baseUrl = 'http://localhost:5000';
    this.results = {
      timestamp: new Date().toISOString(),
      browser: browserType,
      tests: []
    };
  }

  /**
   * Initialize browser and page for testing
   */
  async init(playwright) {
    console.log(`🚀 Launching ${this.browserType} browser for performance testing`);
    
    this.browser = await playwright[this.browserType].launch({
      headless: false, // Show browser for visual feedback
      args: [
        '--disable-web-security',
        '--disable-features=TranslateUI',
        '--no-sandbox',
        '--disable-dev-shm-usage'
      ]
    });

    const context = await this.browser.newContext({
      viewport: { width: 1920, height: 1080 },
      // Enable performance metrics collection
      recordVideo: { dir: './performance-videos/', size: { width: 1920, height: 1080 } }
    });

    this.page = await context.newPage();

    // Enable performance monitoring
    await this.page.addInitScript(() => {
      // Inject performance monitoring into the page
      window.__PERFORMANCE_METRICS__ = [];
      
      // Monitor long tasks
      if ('PerformanceObserver' in window) {
        const observer = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            window.__PERFORMANCE_METRICS__.push({
              type: 'longtask',
              duration: entry.duration,
              startTime: entry.startTime
            });
          }
        });
        try {
          observer.observe({ entryTypes: ['longtask'] });
        } catch (e) {
          console.log('Long task observer not supported');
        }
      }
    });

    console.log('✅ Browser initialized with performance monitoring');
    return this;
  }

  /**
   * Load test scripts into the page
   */
  async loadTestScripts() {
    console.log('📦 Loading performance test scripts');
    
    const performanceAnalyzer = await fs.readFile('./performance-analysis.js', 'utf8');
    const testDataGenerator = await fs.readFile('./test-data-generator.js', 'utf8');
    
    await this.page.addInitScript(performanceAnalyzer);
    await this.page.addInitScript(testDataGenerator);
    
    console.log('✅ Test scripts loaded');
  }

  /**
   * Navigate to dashboard and wait for load
   */
  async navigateToDashboard() {
    console.log('🔍 Navigating to dashboard');
    
    await this.page.goto(this.baseUrl, { waitUntil: 'networkidle' });
    
    // Wait for dashboard to be fully loaded
    await this.page.waitForSelector('[data-testid="repository-list"], .grid', { 
      timeout: 10000 
    }).catch(() => {
      console.warn('⚠️ Repository list selector not found, continuing with fallback');
    });
    
    // Wait for any loading states to complete
    await this.page.waitForTimeout(2000);
    
    console.log('✅ Dashboard loaded');
  }

  /**
   * Test with different repository counts
   */
  async testWithRepositoryCount(count) {
    console.log(`🧪 Testing performance with ${count} repositories`);
    
    const testStart = Date.now();
    
    try {
      // Inject test data
      await this.page.evaluate((repositoryCount) => {
        const generator = new window.TestDataGenerator();
        return generator.injectTestData(repositoryCount);
      }, count);
      
      // Force a re-render by triggering a state change if possible
      await this.page.evaluate(() => {
        // Try to trigger a refresh if a refresh button exists
        const refreshButton = document.querySelector('button[data-testid="refresh"], button:has-text("Refresh")');
        if (refreshButton) {
          refreshButton.click();
        } else {
          // Force a page refresh to load new data
          window.location.reload();
        }
      });
      
      // Wait for the new data to render
      await this.page.waitForTimeout(3000);
      
      // Measure Core Web Vitals
      const webVitals = await this.measureWebVitals();
      
      // Measure scroll performance
      const scrollMetrics = await this.measureScrollPerformance();
      
      // Measure memory usage
      const memoryMetrics = await this.measureMemoryUsage();
      
      // Capture performance timeline
      const performanceEntries = await this.capturePerformanceEntries();
      
      const testResult = {
        repositoryCount: count,
        duration: Date.now() - testStart,
        webVitals,
        scrollMetrics,
        memoryMetrics,
        performanceEntries,
        timestamp: new Date().toISOString()
      };
      
      this.results.tests.push(testResult);
      
      console.log(`✅ Test completed for ${count} repositories`);
      console.log(`   - Scroll FPS: ${scrollMetrics.averageFps}`);
      console.log(`   - Memory: ${memoryMetrics.heapUsed}MB`);
      console.log(`   - LCP: ${webVitals.lcp}ms`);
      
      return testResult;
      
    } catch (error) {
      console.error(`❌ Test failed for ${count} repositories:`, error);
      return { 
        repositoryCount: count, 
        error: error.message,
        duration: Date.now() - testStart
      };
    }
  }

  /**
   * Measure Core Web Vitals
   */
  async measureWebVitals() {
    return await this.page.evaluate(() => {
      return new Promise((resolve) => {
        const vitals = {};
        
        // Largest Contentful Paint
        new PerformanceObserver((entryList) => {
          const entries = entryList.getEntries();
          vitals.lcp = entries[entries.length - 1].startTime;
        }).observe({ entryTypes: ['largest-contentful-paint'] });
        
        // First Input Delay
        new PerformanceObserver((entryList) => {
          const entries = entryList.getEntries();
          vitals.fid = entries[0].processingStart - entries[0].startTime;
        }).observe({ entryTypes: ['first-input'] });
        
        // Cumulative Layout Shift
        let clsValue = 0;
        new PerformanceObserver((entryList) => {
          for (const entry of entryList.getEntries()) {
            if (!entry.hadRecentInput) {
              clsValue += entry.value;
            }
          }
          vitals.cls = clsValue;
        }).observe({ entryTypes: ['layout-shift'] });
        
        // Wait a bit to collect metrics
        setTimeout(() => {
          resolve({
            lcp: Math.round(vitals.lcp || 0),
            fid: Math.round(vitals.fid || 0),
            cls: Math.round((vitals.cls || 0) * 1000) / 1000
          });
        }, 2000);
      });
    });
  }

  /**
   * Measure scroll performance by simulating scroll events
   */
  async measureScrollPerformance() {
    return await this.page.evaluate(() => {
      return new Promise((resolve) => {
        const frameStats = {
          frames: 0,
          frameRates: [],
          droppedFrames: 0
        };
        
        let lastFrameTime = performance.now();
        const duration = 3000; // 3 seconds of measurement
        const startTime = performance.now();
        
        const measureFrame = (currentTime) => {
          const frameDuration = currentTime - lastFrameTime;
          const fps = 1000 / frameDuration;
          
          frameStats.frames++;
          frameStats.frameRates.push(fps);
          
          if (fps < 50) {
            frameStats.droppedFrames++;
          }
          
          lastFrameTime = currentTime;
          
          if (currentTime - startTime < duration) {
            requestAnimationFrame(measureFrame);
          } else {
            const avgFps = frameStats.frameRates.reduce((a, b) => a + b, 0) / frameStats.frameRates.length;
            const minFps = Math.min(...frameStats.frameRates);
            const dropRate = (frameStats.droppedFrames / frameStats.frames) * 100;
            
            resolve({
              averageFps: Math.round(avgFps * 10) / 10,
              minimumFps: Math.round(minFps * 10) / 10,
              droppedFramePercentage: Math.round(dropRate * 10) / 10,
              totalFrames: frameStats.frames
            });
          }
        };
        
        // Start scroll simulation
        const scrollContainer = document.documentElement;
        const maxScroll = scrollContainer.scrollHeight - scrollContainer.clientHeight;
        let scrollDirection = 1;
        let scrollPosition = 0;
        
        const scrollInterval = setInterval(() => {
          scrollPosition += scrollDirection * 10;
          if (scrollPosition >= maxScroll || scrollPosition <= 0) {
            scrollDirection *= -1;
          }
          scrollContainer.scrollTop = scrollPosition;
        }, 16);
        
        // Start frame measurement
        requestAnimationFrame(measureFrame);
        
        // Stop scrolling after duration
        setTimeout(() => {
          clearInterval(scrollInterval);
        }, duration);
      });
    });
  }

  /**
   * Measure memory usage
   */
  async measureMemoryUsage() {
    return await this.page.evaluate(() => {
      if (performance.memory) {
        return {
          heapUsed: Math.round(performance.memory.usedJSHeapSize / 1048576),
          heapTotal: Math.round(performance.memory.totalJSHeapSize / 1048576),
          heapLimit: Math.round(performance.memory.jsHeapSizeLimit / 1048576)
        };
      }
      return { error: 'Memory API not available' };
    });
  }

  /**
   * Capture performance timeline entries
   */
  async capturePerformanceEntries() {
    return await this.page.evaluate(() => {
      const entries = performance.getEntriesByType('measure')
        .concat(performance.getEntriesByType('navigation'))
        .concat(performance.getEntriesByType('paint'));
      
      return entries.map(entry => ({
        name: entry.name,
        type: entry.entryType,
        startTime: Math.round(entry.startTime),
        duration: Math.round(entry.duration || 0)
      }));
    });
  }

  /**
   * Run comprehensive performance test suite
   */
  async runTestSuite() {
    console.log('🧪 Starting comprehensive dashboard performance test suite');
    
    const testCounts = [10, 50, 100, 200, 500];
    
    for (const count of testCounts) {
      await this.testWithRepositoryCount(count);
      
      // Small delay between tests
      await this.page.waitForTimeout(1000);
    }
    
    // Generate summary report
    this.generateSummaryReport();
    
    console.log('✅ Performance test suite completed');
    return this.results;
  }

  /**
   * Generate summary report with recommendations
   */
  generateSummaryReport() {
    const summary = {
      testCount: this.results.tests.length,
      avgScrollFps: 0,
      minScrollFps: 100,
      maxMemoryUsage: 0,
      recommendations: []
    };

    // Analyze results
    this.results.tests.forEach(test => {
      if (test.error) return;
      
      summary.avgScrollFps += test.scrollMetrics.averageFps;
      summary.minScrollFps = Math.min(summary.minScrollFps, test.scrollMetrics.minimumFps);
      
      if (test.memoryMetrics.heapUsed) {
        summary.maxMemoryUsage = Math.max(summary.maxMemoryUsage, test.memoryMetrics.heapUsed);
      }
    });

    const validTests = this.results.tests.filter(t => !t.error);
    summary.avgScrollFps = Math.round(summary.avgScrollFps / validTests.length * 10) / 10;

    // Generate recommendations
    if (summary.minScrollFps < 60) {
      summary.recommendations.push({
        type: 'performance',
        priority: 'high',
        message: `Minimum scroll FPS (${summary.minScrollFps}) below 60fps target. Implement virtual scrolling.`
      });
    }

    if (summary.maxMemoryUsage > 100) {
      summary.recommendations.push({
        type: 'memory',
        priority: 'medium',
        message: `High memory usage detected (${summary.maxMemoryUsage}MB). Review for memory leaks.`
      });
    }

    this.results.summary = summary;
  }

  /**
   * Export results to JSON file
   */
  async saveResults(filename) {
    const outputPath = path.join(__dirname, filename || `performance-test-results-${Date.now()}.json`);
    await fs.writeFile(outputPath, JSON.stringify(this.results, null, 2));
    console.log(`💾 Results saved to: ${outputPath}`);
    return outputPath;
  }

  /**
   * Clean up browser resources
   */
  async cleanup() {
    if (this.browser) {
      await this.browser.close();
      console.log('🧹 Browser closed');
    }
  }
}

module.exports = DashboardPerformanceTest;

// CLI usage if run directly
if (require.main === module) {
  const { chromium } = require('playwright');
  
  async function runTests() {
    const tester = new DashboardPerformanceTest('chromium');
    
    try {
      await tester.init({ chromium });
      await tester.loadTestScripts();
      await tester.navigateToDashboard();
      
      const results = await tester.runTestSuite();
      await tester.saveResults();
      
      console.log('📊 Performance Test Summary:');
      console.log(`   - Average Scroll FPS: ${results.summary.avgScrollFps}`);
      console.log(`   - Minimum Scroll FPS: ${results.summary.minScrollFps}`);
      console.log(`   - Max Memory Usage: ${results.summary.maxMemoryUsage}MB`);
      console.log(`   - Recommendations: ${results.summary.recommendations.length}`);
      
    } catch (error) {
      console.error('❌ Test suite failed:', error);
    } finally {
      await tester.cleanup();
    }
  }
  
  runTests();
}