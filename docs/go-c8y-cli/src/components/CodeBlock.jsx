/* eslint react/jsx-key: 0 */
import React from 'react';
import Highlight, { defaultProps } from 'prism-react-renderer'
import { LiveProvider, LiveEditor, LiveError, LivePreview } from 'react-live'
import { mdx } from '@mdx-js/react';
import useThemeContext from '@theme/hooks/useThemeContext';
import lightTheme from 'prism-react-renderer/themes/github';
import darkTheme from 'prism-react-renderer/themes/dracula';


const c8yCommands = {
    // alarms
    'c8y alarms create': 'New-Alarm',
    'c8y alarms get': 'Get-Alarm',
    'c8y alarms delete': 'Remove-Alarm',
    'c8y alarms subscribe': 'Watch-Alarm',
    'c8y alarms list': 'New-AlarmCollection',

    // events
    'c8y events create': 'New-Event',
    'c8y events get': 'Get-Event',
    'c8y events delete': 'Remove-Event',
    'c8y events subscribe': 'Watch-Event',
    'c8y events list': 'New-EventCollection',

    'c8y devices list': 'New-DeviceCollection',
};

function getCmdlet(parts) {
    let cmdlet = '';
    let prefix = [];
    let lastIdx = 0;
    for (let i = 0; i < parts.length; i++) {
        if (parts[i].startsWith('-')) {
            break;
        }
        prefix.push(parts[i]);
        lastIdx++;
    }
    
    if (parts.length > 2) {
        cmdlet = c8yCommands[prefix.join(' ')]
    }

    if (!cmdlet) {
        return parts
    }
    return [cmdlet, ...parts.slice(lastIdx)]
}

function transformToPowerShell(code = '') {
    let parts = code.split(' ');

    parts = getCmdlet(parts);

    if (parts.length) {
        for (let i = 0; i < parts.length; i++) {
            if (parts[i].startsWith('--')) {
                parts[i] = '-' + parts[i].substr(2, 1).toUpperCase() + parts[i].substr(3)
            } else if (parts[i].startsWith('-')) {
            }
        }
    }
    return parts.join(' ').replace('\\', '`');
}


export default ({ children, className = 'bash', live = false, render = false, transform = false }) => {
    const { isDarkTheme } = useThemeContext();

    const language = className && className.replace ? className.replace(/language-/, '') : 'bash';

    if (live) {
        return (
            <div style={{ marginTop: '40px' }}>
                <LiveProvider
                    code={children.trim()}
                    transformCode={code => '/** @jsx mdx */' + code}
                    scope={{ mdx }}
                >
                    <LivePreview />
                    <LiveEditor />
                    <LiveError />
                </LiveProvider>
            </div>
        )
    }

    if (render) {
        return (
            <div style={{ marginTop: '40px' }}>
                <LiveProvider code={children}>
                    <LivePreview />
                </LiveProvider>
            </div>
        )
    }
    let childrenCode = '';
    if (children && typeof children.trim == 'function') {
        if (transform) {
            childrenCode = transformToPowerShell(children.trim());
        } else {
            childrenCode = children.trim();
        }
    }

    const theme = isDarkTheme ? darkTheme : lightTheme;

    return (
        <Highlight {...defaultProps} theme={theme} code={childrenCode} language={language}>
            {({ className, style, tokens, getLineProps, getTokenProps }) => (
                <pre className={className} style={{ ...style, padding: '20px' }}>
                    {tokens.map((line, i) => (
                        <div key={i} {...getLineProps({ line, key: i })}>
                            {line.map((token, key) => (
                                <span key={key} {...getTokenProps({ token, key })} />
                            ))}
                        </div>
                    ))}
                </pre>
            )}
        </Highlight>
    )
}
