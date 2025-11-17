import React, { useEffect, useRef, useState, useCallback } from 'react';
import { Terminal as XTerm } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { Button, Space } from 'antd';
import { ReloadOutlined, FullscreenOutlined, FullscreenExitOutlined } from '@ant-design/icons';
import 'xterm/css/xterm.css';

interface TerminalProps {
  endpoint: string;
  workerId: string;
  onClose?: () => void;
}

const Terminal: React.FC<TerminalProps> = ({ endpoint, workerId, onClose: _onClose }) => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<XTerm | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [reconnectKey, setReconnectKey] = useState(0);

  const handleReconnect = useCallback(() => {
    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }
    if (xtermRef.current) {
      xtermRef.current.dispose();
    }
    // Trigger re-initialization
    setReconnectKey((prev) => prev + 1);
  }, []);

  const handleFullscreen = useCallback(() => {
    if (!containerRef.current) return;

    if (!isFullscreen) {
      if (containerRef.current.requestFullscreen) {
        containerRef.current.requestFullscreen();
      }
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen();
      }
    }
  }, [isFullscreen]);

  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
      // Resize terminal when entering/exiting fullscreen
      setTimeout(() => {
        if (fitAddonRef.current) {
          fitAddonRef.current.fit();
        }
      }, 100);
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
    };
  }, []);

  useEffect(() => {
    console.log('Terminal component mounted', { endpoint, workerId, reconnectKey });
    if (!terminalRef.current) return;

    // Create terminal
    const term = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
      },
      rows: 30,
      cols: 100,
    });

    // Create fit addon
    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    // Open terminal in DOM
    term.open(terminalRef.current);
    fitAddon.fit();

    xtermRef.current = term;
    fitAddonRef.current = fitAddon;

    // Connect to WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/endpoints/${endpoint}/workers/exec?worker_id=${workerId}`;

    console.log('Connecting to WebSocket:', wsUrl);
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      term.writeln('\r\n\x1b[32mConnected to worker ' + workerId + '\x1b[0m\r\n');

      // Send terminal input to WebSocket
      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(data);
        }
      });
    };

    ws.onmessage = (event) => {
      if (event.data instanceof Blob) {
        // Binary message
        const reader = new FileReader();
        reader.onload = () => {
          const text = reader.result as string;
          term.write(text);
        };
        reader.readAsText(event.data);
      } else {
        // Text message
        term.write(event.data);
      }
    };

    ws.onerror = (error) => {
      term.writeln('\r\n\x1b[31mWebSocket error occurred\x1b[0m\r\n');
      console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
      term.writeln('\r\n\x1b[33mConnection closed\x1b[0m\r\n');
    };

    // Handle window resize
    const handleResize = () => {
      if (fitAddonRef.current) {
        fitAddonRef.current.fit();
      }
    };
    window.addEventListener('resize', handleResize);

    // Cleanup
    return () => {
      window.removeEventListener('resize', handleResize);
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (xtermRef.current) {
        xtermRef.current.dispose();
      }
    };
  }, [endpoint, workerId, reconnectKey]);

  return (
    <div
      ref={containerRef}
      style={{
        height: '100%',
        width: '100%',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: '#1e1e1e',
      }}
    >
      {/* Toolbar */}
      <div
        style={{
          padding: '8px 12px',
          backgroundColor: '#2d2d2d',
          borderBottom: '1px solid #404040',
          display: 'flex',
          justifyContent: 'flex-end',
        }}
      >
        <Space size="small">
          <Button
            size="small"
            icon={<ReloadOutlined />}
            onClick={handleReconnect}
            style={{ color: '#d4d4d4', borderColor: '#404040' }}
          >
            Reconnect
          </Button>
          <Button
            size="small"
            icon={isFullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
            onClick={handleFullscreen}
            style={{ color: '#d4d4d4', borderColor: '#404040' }}
          >
            {isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'}
          </Button>
        </Space>
      </div>

      {/* Terminal */}
      <div
        ref={terminalRef}
        style={{
          flex: 1,
          padding: '10px',
          backgroundColor: '#1e1e1e',
          overflow: 'hidden',
        }}
      />
    </div>
  );
};

export default Terminal;
