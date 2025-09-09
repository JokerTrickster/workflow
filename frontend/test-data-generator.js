/**
 * Test Repository Data Generator
 * Generates realistic repository data for performance testing
 */

class TestDataGenerator {
  constructor() {
    this.languages = [
      'JavaScript', 'TypeScript', 'Python', 'Java', 'C++', 'Go', 'Rust', 'C#',
      'Ruby', 'PHP', 'Kotlin', 'Swift', 'Scala', 'Dart', 'R', 'MATLAB'
    ];
    
    this.frameworks = [
      'React', 'Vue.js', 'Angular', 'Express.js', 'Django', 'Flask', 'Spring Boot',
      'Laravel', 'Ruby on Rails', 'ASP.NET', 'FastAPI', 'Gin', 'Actix', 'Phoenix'
    ];
    
    this.projectTypes = [
      'Web Application', 'API Service', 'Mobile App', 'Library', 'CLI Tool',
      'Desktop App', 'Game', 'Bot', 'Data Analysis', 'Machine Learning',
      'DevOps Tool', 'Blockchain', 'IoT Project', 'Frontend Component'
    ];
    
    this.adjectives = [
      'awesome', 'modern', 'fast', 'secure', 'scalable', 'lightweight', 'robust',
      'flexible', 'powerful', 'efficient', 'innovative', 'advanced', 'smart',
      'clean', 'minimal', 'elegant', 'comprehensive', 'versatile', 'reliable'
    ];
    
    this.nouns = [
      'dashboard', 'platform', 'framework', 'toolkit', 'manager', 'builder',
      'generator', 'analyzer', 'optimizer', 'monitor', 'tracker', 'processor',
      'parser', 'validator', 'converter', 'transformer', 'collector', 'scheduler'
    ];
  }

  /**
   * Generate a random repository name
   */
  generateRepositoryName() {
    const adjective = this.randomChoice(this.adjectives);
    const noun = this.randomChoice(this.nouns);
    const suffix = Math.random() > 0.7 ? `-${this.randomChoice(['app', 'api', 'cli', 'ui', 'core'])}` : '';
    return `${adjective}-${noun}${suffix}`;
  }

  /**
   * Generate a realistic repository description
   */
  generateDescription() {
    const projectType = this.randomChoice(this.projectTypes);
    const framework = this.randomChoice(this.frameworks);
    const adjective = this.randomChoice(this.adjectives);
    
    const templates = [
      `A ${adjective} ${projectType.toLowerCase()} built with ${framework}`,
      `${projectType} for building ${adjective} applications`,
      `${framework}-based ${projectType.toLowerCase()} with ${adjective} features`,
      `${adjective.charAt(0).toUpperCase() + adjective.slice(1)} ${projectType.toLowerCase()} powered by ${framework}`,
      `Open-source ${projectType.toLowerCase()} designed for ${adjective} performance`
    ];
    
    return this.randomChoice(templates);
  }

  /**
   * Generate realistic repository statistics
   */
  generateStats() {
    // Simulate realistic distribution of repository popularity
    const popularity = Math.random();
    let stars, forks;
    
    if (popularity < 0.7) {
      // Most repos have low stars
      stars = Math.floor(Math.random() * 50);
      forks = Math.floor(stars * (0.1 + Math.random() * 0.3));
    } else if (popularity < 0.9) {
      // Some repos are moderately popular
      stars = Math.floor(50 + Math.random() * 200);
      forks = Math.floor(stars * (0.15 + Math.random() * 0.25));
    } else {
      // Few repos are very popular
      stars = Math.floor(250 + Math.random() * 2000);
      forks = Math.floor(stars * (0.2 + Math.random() * 0.3));
    }
    
    return { stars, forks };
  }

  /**
   * Generate a realistic update date
   */
  generateUpdateDate() {
    const now = Date.now();
    const daysAgo = Math.floor(Math.random() * 365); // Within last year
    return new Date(now - (daysAgo * 24 * 60 * 60 * 1000)).toISOString();
  }

  /**
   * Get a random choice from an array
   */
  randomChoice(array) {
    return array[Math.floor(Math.random() * array.length)];
  }

  /**
   * Generate a single repository object
   */
  generateRepository(id) {
    const name = this.generateRepositoryName();
    const language = this.randomChoice(this.languages);
    const { stars, forks } = this.generateStats();
    const isPrivate = Math.random() < 0.3; // 30% chance of being private
    const isConnected = Math.random() < 0.4; // 40% chance of being connected
    
    return {
      id,
      name,
      full_name: `testuser/${name}`,
      description: this.generateDescription(),
      language,
      stargazers_count: stars,
      forks_count: forks,
      updated_at: this.generateUpdateDate(),
      html_url: `https://github.com/testuser/${name}`,
      private: isPrivate,
      is_connected: isConnected,
      owner: {
        login: 'testuser',
        avatar_url: 'https://avatars.githubusercontent.com/u/1?v=4'
      }
    };
  }

  /**
   * Generate an array of repositories
   */
  generateRepositories(count) {
    console.log(`🏭 Generating ${count} test repositories...`);
    const repositories = [];
    
    for (let i = 1; i <= count; i++) {
      repositories.push(this.generateRepository(i));
    }
    
    console.log(`✅ Generated ${repositories.length} repositories`);
    return repositories;
  }

  /**
   * Inject test data into the application (override useRepositories hook)
   */
  injectTestData(count) {
    const repositories = this.generateRepositories(count);
    
    // Store in window for access by the application
    window.__TEST_REPOSITORIES__ = repositories;
    
    // Try to find and override the repositories hook if possible
    if (window.React && window.React.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED) {
      console.log('⚛️ Attempting to inject test data into React application');
      
      // This is a development-only hack for testing
      const originalHook = window.__REPOSITORY_HOOK_OVERRIDE__;
      window.__REPOSITORY_HOOK_OVERRIDE__ = () => ({
        repositories,
        isLoading: false,
        error: null,
        refetch: () => Promise.resolve(),
        isRefetching: false,
        connectRepository: (id) => {
          const repo = repositories.find(r => r.id === id);
          if (repo) repo.is_connected = true;
          return Promise.resolve();
        }
      });
      
      console.log('✅ Test data injected successfully');
    }
    
    return repositories;
  }

  /**
   * Create performance test scenarios with different data sizes
   */
  createTestScenarios() {
    const scenarios = [
      { name: 'Small Dataset', count: 10, description: 'Baseline performance with few repositories' },
      { name: 'Medium Dataset', count: 50, description: 'Typical user with moderate repository count' },
      { name: 'Large Dataset', count: 100, description: 'Target scenario for 60fps performance' },
      { name: 'Extra Large Dataset', count: 200, description: 'Stress test with many repositories' },
      { name: 'Extreme Dataset', count: 500, description: 'Maximum load test scenario' }
    ];
    
    return scenarios.map(scenario => ({
      ...scenario,
      data: this.generateRepositories(scenario.count),
      testFunction: (analyzer) => {
        this.injectTestData(scenario.count);
        return analyzer.runFullAnalysis(scenario.count);
      }
    }));
  }

  /**
   * Export test data to JSON file
   */
  exportTestData(count, filename) {
    const repositories = this.generateRepositories(count);
    const data = {
      metadata: {
        count,
        generated: new Date().toISOString(),
        generator: 'TestDataGenerator v1.0'
      },
      repositories
    };
    
    const blob = new Blob([JSON.stringify(data, null, 2)], {
      type: 'application/json'
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename || `test-repositories-${count}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    console.log(`💾 Exported ${count} repositories to ${filename}`);
  }
}

// Export for use in browser console or testing
window.TestDataGenerator = TestDataGenerator;

// Auto-initialize if script is loaded directly
if (typeof module === 'undefined') {
  console.log('🏭 Test Data Generator loaded');
  console.log('Run: new TestDataGenerator().injectTestData(100)');
}

// Usage examples:
// const generator = new TestDataGenerator();
// generator.injectTestData(100); // Inject 100 test repositories
// generator.exportTestData(200, 'large-test-set.json'); // Export 200 repos to file