/* eslint react/jsx-key: 0 */
import React from 'react';
import { LiveProvider, LiveEditor, LiveError, LivePreview } from 'react-live';
import { mdx } from '@mdx-js/react';
import { useColorMode } from '@docusaurus/theme-common';
import lightTheme from 'prism-react-renderer/themes/github';
import darkTheme from 'prism-react-renderer/themes/dracula';
import Code from '@docusaurus/theme-classic/lib/theme/CodeBlock';


const c8yCommands = {
    // alarms
    'c8y alarms create': 'New-Alarm',
    'c8y alarms get': 'Get-Alarm',
    'c8y alarms delete': 'Remove-Alarm',
    'c8y alarms subscribe': 'Watch-Alarm',
    'c8y alarms list': 'Get-AlarmCollection',

    // events
    'c8y events create': 'New-Event',
    'c8y events get': 'Get-Event',
    'c8y events delete': 'Remove-Event',
    'c8y events subscribe': 'Watch-Event',
    'c8y events list': 'Get-EventCollection',

    // measurements
    'c8y measurements create': 'New-Measurement',

    // operations
    'c8y operations create': 'New-Operation',
    'c8y operations get': 'Get-Operation',
    'c8y operations update': 'Update-Operation',
    'c8y operations delete': 'Remove-Operation',
    'c8y operations subscribe': 'Watch-Operation',
    'c8y operations list': 'Get-OperationCollection',

    // auditrecords
    'c8y auditrecords create': 'New-AuditRecord',
    'c8y auditrecords get': 'Get-AuditRecord',
    'c8y auditrecords delete': 'Remove-AuditRecord',
    'c8y auditrecords subscribe': 'Watch-AuditRecord',
    'c8y auditrecords list': 'Get-AuditRecordCollection',

    'c8y devices create': 'New-Device',
    'c8y devices list': 'New-DeviceCollection',
    'c8y devices get': 'Get-Device',
    'c8y devices update': 'Update-Device',
    'c8y devices delete': 'Remove-Device',
    'c8y devices setRequiredAvailability': 'Set-DeviceRequiredAvailability',

    'c8y inventory create': 'New-ManagedObject',
    'c8y inventory get': 'Get-ManagedObject',
    'c8y inventory update': 'Update-ManagedObject',
    'c8y inventory list': 'Get-ManagedObjectCollection',
    'c8y inventory find': 'Find-ManagedObjectCollection',
    'c8y inventory findByText': 'Find-ByTextManagedObjectCollection',

    'c8y devicegroups create': 'New-DeviceGroup',
    'c8y devicegroups list': 'Get-DeviceGroupCollection',
    'c8y devicegroups assignDevice': 'Add-DeviceToGroup',
    'c8y devicegroups unassignDevice': 'Remove-DeviceFromGroup',
    'c8y devicegroups listAssets': 'Get-DeviceGroupChildAssetCollection',
    
    'c8y applications create': 'New-Application',
    'c8y applications createHostedApplication': 'New-HostedApplication',
    'c8y applications list': 'Get-ApplicationCollection',
    'c8y applications get': 'Get-Application',
    'c8y applications update': 'Update-Application',
    'c8y applications delete': 'Remove-Application',

    // Software
    'c8y software versions install': 'Install-SoftwareVersion',
    'c8y software versions uninstall': 'Remove-SoftwareVersion',

    // binaries
    'c8y binaries get': 'Get-Binary',

    // microservices
    'c8y microservices get': 'Get-Microservice',
    'c8y microservices list': 'Get-MicroserviceCollection',
    'c8y microservices update': 'Update-Microservice',
    'c8y microservices delete': 'Delete-Microservice',
    'c8y microservices create': 'New-Microservice',
    'c8y microservices enable': 'Enable-Microservice',
    'c8y microservices disable': 'Disable-Microservice',
    'c8y microservices getBootstrapUser': 'Get-MicroserviceBootstrapUser',

    'c8y template execute': 'Invoke-Template',
    
    'c8y sessions create': 'New-Session',
    'c8y currentuser get': 'Get-CurrentUser',
    'set-session': 'Set-Session',
    
    'cat ': 'Get-Content ',
    '-o csv': '--output csv',
    '-o json': '--output json',
    '-p ': '--pageSize ',
    
};

function replaceAll(string, search, replace) {
    return string.split(search).join(replace);
}

function convertToCmdlets(code = '') {
    const keys = Object.keys(c8yCommands);
    for (let index = 0; index < keys.length; index++) {
        const element = c8yCommands[keys[index]];
        code = replaceAll(code, keys[index], element);
    }
    code = code.replace(/\\/g, '`');
    return code;
}

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
    let parts = convertToCmdlets(code).split(' ');

    // parts = getCmdlet(parts);

    if (parts.length) {
        for (let i = 0; i < parts.length; i++) {
            if (parts[i].startsWith('--')) {
                parts[i] = '-' + parts[i].substr(2, 1).toUpperCase() + parts[i].substr(3)
            } else if (parts[i].startsWith('-')) {
            }
        }
    }
    return parts.join(' ');
}


export default ({ children, className = 'bash', live = false, render = false, transform = false }) => {
    const { colorMode } = useColorMode();
    const isDarkTheme = colorMode === "dark";
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
        if (`${transform}` == 'true') {
            childrenCode = transformToPowerShell(children.trim());
        } else {
            childrenCode = children.trim();
            if (className === 'powershell') {
                childrenCode = childrenCode.replace(/\\/g, '`')
            }
        }
    }

    const theme = isDarkTheme ? darkTheme : lightTheme;
    return (
        <Code className={className} theme={theme}>
            {childrenCode}
        </Code>
    )
}
