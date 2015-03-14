/// rxjs websocket

var HOST = 'ws://127.0.0.1:6201';
var log = console.log;

var openingObserver = Rx.Observer.create(function() { console.log('Opening socket'); });
var closingObserver = Rx.Observer.create(function() { console.log('Closing socket'); });

var stateSocket = Rx.DOM.fromWebSocket(
  HOST +'/node/state', null, openingObserver, closingObserver);
var RxNodeState = stateSocket.map(function(e){ return JSON.parse(e.data); });

var migrateSocket = Rx.DOM.fromWebSocket(
  HOST +'/migrate/state', null, openingObserver, closingObserver);
var RxMigration = migrateSocket.map(function(e){ return JSON.parse(e.data); });

/// react

var MigratingRow = React.createClass({
  render: function() {
    var obj = this.props.obj;
    return (
      <tr className="nodeRow">
        <td>{obj.left}-{obj.right}</td>
        <td>{obj.state}</td>
      </tr>
    );
  }
});

var MigrationTable = React.createClass({
  render: function() {
    var mig = this.props.mig;
    var name = mig.SourceId.substring(0,6)+' to '+mig.TargetId.substring(0,6);
    var rows = mig.Ranges.map(function(range, idx){
      var obj = {left: range.Left, right: range.Right};
      if (idx > mig.CurrRangeIndex)
        obj.state = "Todo";
      if (idx < mig.CurrRangeIndex)
        obj.state = "Done";
      if (idx == mig.CurrRangeIndex)
        obj.state = mig.State
      return <MigratingRow obj={obj} />;
    });
    return (
      <div className="migrateTable">
        <table className="nodeTable">
        <tr className="nodeRow">
        <td>{name}</td>
        <td>{mig.CurrSlot}</td>
        </tr>
        {rows}
      </table>
      </div>
    );
  }
});

var MigrationPanel = React.createClass({
  render: function() {
    var migMap = this.props.migMap;
    var keys = _.keys(migMap).sort();
    var migs = keys.map(function (key) {
      return (
        <MigrationTable mig={migMap[key]} />
      );
    });
    return (
      <div className="migrationPanel">
        {migs}
      </div>
    );
  }
});

/// NodeRangeState

var NodeRangeBarItem = React.createClass({
  render: function() {
    var width = 1024;
    var range = this.props.range;
    var style = {
      left: range.Left*width/16384,
      width: (range.Right-range.Left+1)*width/16384,
      backgroundColor: "#00BB00"
    };
    return (
        <div className="nodeRangeBarItem" style={style}>
        </div>
    );
  }
});

var NodeRangeRow = React.createClass({
  render: function() {
    var id = this.props.nodeid;
    var ranges = this.props.ranges;
    var items = ranges.map(function (range) {
      return (
          <NodeRangeBarItem range={range} />
      );
    });
    return (
      <tr className="nodeRow">
        <td>{id.substring(0,6)}</td>
        <td>
          <div className="nodeRangeBar">
          {items}
          </div>
        </td>
      </tr>
    );
  }
});

var NodeRangeTable = React.createClass({
  render: function() {
    var nodes = this.props.nodes;
    var keys = _.keys(this.props.nodes).filter(function(key){
      return nodes[key].Role == "master";
    }).sort();
    var rows = keys.map(function (key) {
      return (
        <NodeRangeRow nodeid={nodes[key].Id} ranges={nodes[key].Ranges} />
      );
    });
    return (
      <table className="nodeTable">
        {rows}
      </table>
    )
  }
});

/// NodeStateTable

var NodeStateRow = React.createClass({
  render: function() {
    var node = this.props.node;
    var FAIL = node.Fail ? "FAIL":"OK";
    var READ = node.Readable ? "Read":"-";
    var WRITE = node.Writable ? "Write":"-";
    var MIGRATING = node.Migrating ? "Migrating":"-";
    return (
        <tr className="nodeRow">
          <td>{node.State}</td>
          <td>{node.Region}</td>
          <td className={FAIL}>{FAIL}</td>
          <td>{READ}</td>
          <td>{WRITE}</td>
          <td>{node.Role}</td>
          <td>{node.Ip}:{node.Port}</td>
          <td>{node.Id.substring(0,6)}</td>
          <td>{MIGRATING}</td>
          <td>{node.Version}</td>
        </tr>
    );
  }
});

var NodeStateTable = React.createClass({
  render: function() {
    var props = this.props;
    var keys = _.keys(props.nodes).sort();
    var nodes = keys.map(function (key) {
      return (
          <NodeStateRow node={props.nodes[key]} />
      );
    });
    return (
        <table className="nodeTable">
          {nodes}
        </table>
    );
  }
});

var Main = React.createClass({
  componentDidMount: function() {
    var self = this;
    // 也许该用rx-react之类的Addon
    RxNodeState.subscribe(
      function (obj) {
        var nodes = self.props.nodes;
        nodes[obj.Id] = obj;
        self.setState({nodes: nodes});
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });

    RxMigration.subscribe(
      function (obj) {
        var migMap = self.props.migMap;
        migMap[obj.SourceId] = obj;
        self.setState({migMap: migMap});
      },
      function (e) {
        console.log('Error: ', e);
      },
      function (){
        console.log('Closed');
      });
  },
  render: function() {
    var nodes = this.props.nodes;
    var migMap = this.props.migMap;
    return (
      <div className="Main">
        <NodeStateTable nodes={nodes} />
        <NodeRangeTable nodes={nodes} />
        <MigrationPanel migMap={migMap} />
      </div>
    );
  }
});

React.render(
    <Main nodes={{}} migMap={{}} />,
    document.getElementById('content')
);
