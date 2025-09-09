/**
 * Simple Performance Test Runner
 * Runs browser-based performance analysis
 */

const { chromium } = require('playwright');
const fs = require('fs').promises;

async function runPerformanceAnalysis() {
  console.log('🚀 Starting Dashboard Performance Analysis');
  
  const browser = await chromium.launch({ 
    headless: false,
    slowMo: 100 // Slow down for better visibility
  });
  
  const context = await browser.newContext({
    viewport: { width: 1920, height: 1080 }
  });
  
  const page = await context.newPage();
  
  try {
    // Navigate to the dashboard
    console.log('📍 Navigating to dashboard at http://localhost:5000');
    await page.goto('http://localhost:5000', { waitUntil: 'networkidle' });
    
    // Wait for the page to load completely
    await page.waitForTimeout(3000);
    
    // Inject our performance analysis scripts
    console.log('💉 Injecting performance analysis tools');
    
    const performanceScript = await fs.readFile('./performance-analysis.js', 'utf8');
    const testDataScript = await fs.readFile('./test-data-generator.js', 'utf8');
    
    await page.addScriptTag({ content: performanceScript });
    await page.addScriptTag({ content: testDataScript });
    
    // Run performance tests with different data sizes
    const testSizes = [10, 50, 100, 200];
    const results = [];
    
    for (const size of testSizes) {
      console.log(`\n🧪 Testing with ${size} repositories`);
      
      const result = await page.evaluate(async (repositoryCount) => {
        // Generate and inject test data
        const generator = new window.TestDataGenerator();
        const repositories = generator.generateRepositories(repositoryCount);
        
        // Store test data globally for potential access by hooks
        window.__TEST_REPOSITORIES__ = repositories;
        
        // Create performance analyzer
        const analyzer = new window.PerformanceAnalyzer();
        
        // Run the analysis
        try {
          const results = await analyzer.runFullAnalysis(repositoryCount);
          return {
            success: true,
            repositoryCount,
            ...results
          };
        } catch (error) {
          return {
            success: false,
            repositoryCount,
            error: error.message
          };
        }
      }, size);
      
      results.push(result);
      
      // Log key metrics
      if (result.success) {
        console.log(`✅ Test completed for ${size} repositories:`);
        if (result.scrollMetrics) {
          console.log(`   📊 Average FPS: ${result.scrollMetrics.averageFps}`);
          console.log(`   📊 Minimum FPS: ${result.scrollMetrics.minimumFps}`);
          console.log(`   📊 Dropped Frames: ${result.scrollMetrics.droppedFramePercentage}%`);
        }
        if (result.memoryMetrics && result.memoryMetrics.usedJSHeapSize) {
          console.log(`   🧠 Memory Usage: ${result.memoryMetrics.usedJSHeapSize}MB`);
        }
        console.log(`   🎯 Target (60fps): ${result.scrollMetrics?.averageFps >= 60 ? '✅ PASSED' : '❌ FAILED'}`);
      } else {
        console.log(`❌ Test failed for ${size} repositories: ${result.error}`);
      }
      
      // Wait between tests
      await page.waitForTimeout(2000);
    }
    
    // Save results to file
    const timestamp = new Date().toISOString().replace(/:/g, '-');
    const filename = `performance-results-${timestamp}.json`;
    
    await fs.writeFile(filename, JSON.stringify({
      timestamp: new Date().toISOString(),
      browser: 'chromium',
      tests: results,
      summary: generateSummary(results)
    }, null, 2));
    
    console.log(`\n💾 Results saved to: ${filename}`);
    
    // Generate final report
    generateReport(results);
    
  } catch (error) {
    console.error('❌ Performance analysis failed:', error);
  } finally {
    await browser.close();
  }
}

function generateSummary(results) {
  const validResults = results.filter(r => r.success && r.scrollMetrics);
  
  if (validResults.length === 0) {
    return { error: 'No valid test results' };
  }
  
  const avgFps = validResults.reduce((sum, r) => sum + r.scrollMetrics.averageFps, 0) / validResults.length;
  const minFps = Math.min(...validResults.map(r => r.scrollMetrics.minimumFps));
  const maxDropRate = Math.max(...validResults.map(r => r.scrollMetrics.droppedFramePercentage));
  
  const targetMet = validResults.filter(r => r.scrollMetrics.averageFps >= 60).length;
  
  return {
    averageFps: Math.round(avgFps * 10) / 10,
    minimumFps: minFps,
    maxDroppedFrameRate: maxDropRate,
    testsPassingTarget: `${targetMet}/${validResults.length}`,
    overallGrade: targetMet === validResults.length ? 'EXCELLENT' : targetMet > validResults.length / 2 ? 'GOOD' : 'NEEDS_IMPROVEMENT'
  };
}

function generateReport(results) {
  console.log('\n📊 ===== DASHBOARD PERFORMANCE ANALYSIS REPORT =====');
  console.log(`🕐 Test Date: ${new Date().toISOString()}`);
  console.log(`🎯 Target: 60fps scroll performance with 100+ repositories`);
  
  const summary = generateSummary(results);
  
  if (summary.error) {
    console.log('❌ No valid test results to report');
    return;
  }
  
  console.log('\n📈 OVERALL PERFORMANCE METRICS:');
  console.log(`   Average FPS across all tests: ${summary.averageFps}`);
  console.log(`   Minimum FPS recorded: ${summary.minimumFps}`);
  console.log(`   Maximum dropped frame rate: ${summary.maxDroppedFrameRate}%`);
  console.log(`   Tests meeting 60fps target: ${summary.testsPassingTarget}`);
  console.log(`   Overall Grade: ${summary.overallGrade}`);
  
  console.log('\n📊 DETAILED TEST RESULTS:');
  results.forEach(result => {
    if (result.success && result.scrollMetrics) {
      const status = result.scrollMetrics.averageFps >= 60 ? '✅' : '❌';
      console.log(`   ${status} ${result.repositoryCount} repos: ${result.scrollMetrics.averageFps}fps avg, ${result.scrollMetrics.minimumFps}fps min`);
    } else {
      console.log(`   ❌ ${result.repositoryCount} repos: FAILED (${result.error})`);
    }
  });
  
  console.log('\n🎯 RECOMMENDATIONS:');
  
  if (summary.averageFps < 60) {
    console.log('   🔧 HIGH PRIORITY: Implement virtual scrolling for repository lists');
    console.log('   🔧 HIGH PRIORITY: Consider lazy loading of repository cards');
  }
  
  if (summary.maxDroppedFrameRate > 10) {
    console.log('   🔧 MEDIUM PRIORITY: Optimize RepositoryCard rendering');
    console.log('   🔧 MEDIUM PRIORITY: Review useMemo and useCallback implementations');
  }
  
  if (summary.overallGrade === 'EXCELLENT') {
    console.log('   🎉 EXCELLENT: Performance targets are being met!');
    console.log('   📈 NEXT STEPS: Consider implementing IntersectionObserver for infinite scroll');
  }
  
  console.log('\n🔍 NEXT STEPS:');
  console.log('   1. Review the generated JSON file for detailed metrics');
  console.log('   2. Implement virtual scrolling if FPS < 60fps');
  console.log('   3. Integrate IntersectionObserver hook for infinite scroll');
  console.log('   4. Set up performance monitoring in CI/CD pipeline');
  
  console.log('\n✅ Performance analysis complete!');
}

// Run the analysis
runPerformanceAnalysis().catch(console.error);