'use client';

import { useState, useEffect } from 'react';
import { ErrorHandler, ErrorType } from '../utils/errorHandler';

interface NetworkStatus {
  isOnline: boolean;
  effectiveType?: string;
  downlink?: number;
  rtt?: number;
}

export function useNetworkStatus() {
  const [networkStatus, setNetworkStatus] = useState<NetworkStatus>({
    isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
  });

  const [offlineError, setOfflineError] = useState<any>(null);

  useEffect(() => {
    const updateNetworkStatus = () => {
      const isOnline = navigator.onLine;
      
      // Network API가 지원되는 경우 추가 정보 수집
      const connection = (navigator as any)?.connection;
      
      const status: NetworkStatus = {
        isOnline,
        effectiveType: connection?.effectiveType,
        downlink: connection?.downlink,
        rtt: connection?.rtt,
      };
      
      setNetworkStatus(status);

      // 오프라인 상태에서 온라인으로 전환 시 에러 상태 초기화
      if (isOnline && offlineError) {
        setOfflineError(null);
      }

      // 오프라인 상태일 때 에러 설정
      if (!isOnline) {
        const error = ErrorHandler.createError(
          ErrorType.NETWORK,
          'NETWORK_OFFLINE',
          'No internet connection',
          { recoverable: true, retryable: true }
        );
        setOfflineError(error);
      }
    };

    // 초기 상태 설정
    updateNetworkStatus();

    // 이벤트 리스너 등록
    window.addEventListener('online', updateNetworkStatus);
    window.addEventListener('offline', updateNetworkStatus);

    // Network API 변경 감지
    const connection = (navigator as any)?.connection;
    if (connection) {
      connection.addEventListener('change', updateNetworkStatus);
    }

    return () => {
      window.removeEventListener('online', updateNetworkStatus);
      window.removeEventListener('offline', updateNetworkStatus);
      
      if (connection) {
        connection.removeEventListener('change', updateNetworkStatus);
      }
    };
  }, [offlineError]);

  const getConnectionQuality = (): 'excellent' | 'good' | 'fair' | 'poor' | 'unknown' => {
    if (!networkStatus.isOnline) return 'poor';
    if (!networkStatus.effectiveType) return 'unknown';

    switch (networkStatus.effectiveType) {
      case '4g':
        return 'excellent';
      case '3g':
        return 'good';
      case '2g':
        return 'fair';
      case 'slow-2g':
        return 'poor';
      default:
        return 'unknown';
    }
  };

  const shouldReduceRequests = (): boolean => {
    const quality = getConnectionQuality();
    return quality === 'poor' || quality === 'fair';
  };

  const getRecommendedRetryDelay = (): number => {
    const quality = getConnectionQuality();
    
    switch (quality) {
      case 'excellent':
        return 1000; // 1초
      case 'good':
        return 2000; // 2초
      case 'fair':
        return 5000; // 5초
      case 'poor':
        return 10000; // 10초
      default:
        return 3000; // 3초 (기본값)
    }
  };

  return {
    networkStatus,
    offlineError,
    isOnline: networkStatus.isOnline,
    connectionQuality: getConnectionQuality(),
    shouldReduceRequests: shouldReduceRequests(),
    recommendedRetryDelay: getRecommendedRetryDelay(),
    clearOfflineError: () => setOfflineError(null),
  };
}